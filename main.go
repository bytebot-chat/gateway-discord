package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/alexliesenfeld/health"
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
	redisUser = flag.String("ruser", "", "Redis username")
	id        = flag.String("id", "discord", "ID to use when publishing messages")
	verbose   = flag.Bool("verbose", false, "Enable verbose logging")
)

// init is run before main() and is used to initialize the logger
func init() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Setup command line parameters
	flag.StringVar(&Token, "t", "", "Bot Token") // Discord bot token
	flag.Parse()                                 // Parse the command line parameters

	// Parse environment variables
	parseEnv()

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

	// Connect to redis
	rdb = redisConnect(*redisAddr, *redisUser, *redisPass, 0, redisCtx)
	if rdb == nil {
		log.Fatal().Msg("Unable to connect to Redis")
		return
	}

	defer rdb.Close()

	// Ping redis one more time to make sure we're connected
	_, err = rdb.Ping(redisCtx).Result()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("func", "main").
			Msg("Error pinging Redis")
	}

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

	// Handle outbound messages in a goroutine

	// Setup a health check endpoint
	checker := health.NewChecker(
		health.WithCacheDuration(1*time.Second),
		health.WithTimeout(10*time.Second),
		// Check the redis connection with a ping
		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "redis",
				Check: func(ctx context.Context) error {
					_, err := rdb.Ping(ctx).Result()
					return err
				},
			}),

		// Test the redis pubsub connection by subscribing and unsubscribing
		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "redis-pubsub",
				Check: func(ctx context.Context) error {
					pubsub := rdb.Subscribe(ctx, "test")
					_, err := pubsub.Receive(ctx)
					if err != nil {
						return err
					}
					err = pubsub.Close()
					return err
				},
			}),

		// Test the Discord connection by sending a message to our own ID
		health.WithPeriodicCheck(
			15*time.Second,
			3*time.Second,
			health.Check{
				Name: "discord",
				Check: func(ctx context.Context) error {
					_, err := dgo.ChannelMessageSend(dgo.State.User.ID, "Test")
					return err
				},
			}),
	)

	// Register the health check endpoint
	http.Handle("/health", health.NewHandler(checker))

	// Start the health check server
	go log.Fatal().Err(http.ListenAndServe(":8080", nil)).Msg("Error starting health check server")

	// for loop to hold the program open until CTRL-C is pressed
	for {
		select {
		case <-cancelChan:
			log.Info().Str("func", "main").Msg("CTRL-C pressed. Exiting.")
			return
		}
	}
}
