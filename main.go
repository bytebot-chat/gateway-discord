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
	Token string // Discord bot token

	// Global stuff for Redis and Ctrl-C
	redisCtx  context.Context
	cancelCtx context.Context
	cancel    context.CancelFunc
	rdb       *redis.Client

	redisAddr = flag.String("redis", "localhost:6379", "Address and port of redis host")
	redisPass = flag.String("rpass", "", "Redis password")
	id        = flag.String("id", "discord", "ID to use when publishing messages")
	inbound   = flag.String("inbound", "discord-inbound", "Pubsub queue to publish inbound messages to")
	outbound  = flag.String("outbound", "discord-outbound", "Pubsub to subscribe to for sending outbound messages. Defaults to being equivalent to `id`")
	verbose   = flag.Bool("verbose", false, "Enable verbose logging")
)

// init is run before main() and is used to initialize the logger
func init() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Setup command line parameters
	flag.StringVar(&Token, "t", "", "Bot Token") // Discord bot token
	flag.Parse()                                 // Parse the command line parameters

	// Set the contexts before we start the program
	redisCtx = context.Background()
	cancelCtx, cancel = context.WithCancel(context.Background())

	// Set the log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func main() {
	// Hold the program open until CTRL-C is pressed
	log.Info().Str("func", "main").Msg("Gateway is now running. Press CTRL-C to exit.")

	defer cancel()

	// Setup channel for capturing CTRL-C
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt)
	defer func() {
		signal.Stop(cancelChan)
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
		log.Fatal().Err(err).Msg("Unable to create Discord session")
		return
	}

	// Test our connection to Discord
	dgo.AddHandlerOnce(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info().
			Str("func", "main").
			Str("id", r.User.ID).
			Str("username", r.User.Username).
			Str("discriminator", r.User.Discriminator).
			Msg("Connected to Discord")
	})

	// Register the messageCreate func as a callback for MessageCreate events.
	dgo.AddHandler(messageCreate)

	// Setup Redis client
	log.Info().
		Str("func", "main").
		Str("redis", *redisAddr).
		Msg("Connecting to Redis")
	rdb = redisConnect(*redisAddr, *redisPass, 0, redisCtx) // Connect to Redis
	if rdb == nil {
		log.Fatal().Msg("Unable to connect to Redis")
		return
	}
	defer rdb.Close()
	log.Info().
		Str("func", "main").
		Str("redis", *redisAddr).
		Msg("Connected to Redis")

	// Open a websocket connection to Discord and begin listening.
	// Error out if we can't connect
	err = dgo.Open()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("func", "main").
			Msg("Error opening connection to Discord")
	}
	defer dgo.Close()

	// Subscribe to the outbound queue
	log.Info().
		Str("func", "main").
		Str("queue", *outbound).
		Msg("Subscribing to outbound queue")

	// Handle outbound messages in a goroutine
	go handleOutbound(*outbound, rdb, dgo, redisCtx)

	// for loop to hold the program open until CTRL-C is pressed
	for {
		select {
		case <-cancelChan:
			log.Info().Str("func", "main").Msg("CTRL-C pressed. Exiting.")
			return
		}
	}
}
