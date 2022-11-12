package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Variables used for command line parameters
var (
	Token string          // Discord bot token
	ctx   context.Context // Redis context
	rdb   *redis.Client   // Redis client

	redisAddr = flag.String("redis", "localhost:6379", "Address and port of redis host")
	redisPass = flag.String("rpass", "", "Redis password")
	id        = flag.String("id", "discord", "ID to use when publishing messages")
	inbound   = flag.String("inbound", "discord-inbound", "Pubsub queue to publish inbound messages to")
	outbound  = flag.String("outbound", *id, "Pubsub to subscribe to for sending outbound messages. Defaults to being equivalent to `id`")
)

// init is run before main() and is used to initialize the logger
func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	flag.StringVar(&Token, "t", "", "Bot Token") // Discord bot token
	flag.Parse()                                 // Parse the command line parameters
}

func main() {
	// Hold the program open until CTRL-C is pressed
	log.Info().Str("func", "main").Msg("Gateway is now running. Press CTRL-C to exit.")

	// Setup context for capturing CTRL-C
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup channel for capturing CTRL-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()

	// Log our input parameters
	log.Info().
		Str("func", "main").
		Str("id", *id).
		Str("inbound", *inbound).
		Str("outbound", *outbound).
		Str("redis", *redisAddr).
		Msg("Starting Discord gateway")

	// Create a new Discordgo session using the provided bot token.
	dgo, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Err(err).
			Str("func", "main").
			Msg("Error creating Discord session")
		return
	}

	r := redisConnect(*redisAddr, *redisPass, 0, ctx) // Connect to Redis

	// Register the messageCreate func as a callback for MessageCreate events.
	dgo.AddHandler(messageCreate)

	// Open a connection to Redis pubsub and subscribe to the inbound topic
	go handleInbound(*inbound, r, dgo)

	// Wait for CTRL-C
	select {
	case <-c: // CTRL-C
		cancel()
	case <-ctx.Done():
	}
	<-c // Wait for second CTRL-C
	os.Exit(1)
}
