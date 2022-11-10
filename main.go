package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bytebot-chat/gateway-discord/model"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
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

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	flag.StringVar(&Token, "t", "", "Bot Token") // Discord bot token
	flag.Parse()                                 // Parse the command line parameters
}

func main() {
	log.Info().
		Str("func", "main").
		Str("id", *id).
		Str("inbound", *inbound).
		Str("outbound", *outbound).
		Str("redis", *redisAddr).
		Msg("Starting Discord gateway")

	rdb = rdbConnect(*redisAddr, *redisPass, 1) // Connect to redis
	ctx = context.Background()                  // Redis context

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Err(err).
			Str("func", "main").
			Msg("Unable to continue without connection. Exiting!")
		os.Exit(1)
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info().
			Str("func", "main").
			Str("id", r.User.ID).
			Str("username", r.User.Username).
			Str("discriminator", r.User.Discriminator).
			Msg("Discord gateway connected")
	})
	// Cleanly close down the Discord session.
	defer dg.Close()
	dg.AddHandler(messageCreate)                                                      // Register the messageCreate func as a callback for MessageCreate events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged) // Listen for all low-privileged intents

	err = dg.Open() // Open the websocket and begin listening.
	if err != nil {
		log.Err(err).
			Str("func", "main").
			Msg("Unable to open channel for reading messages")
		os.Exit(1) // Exit if we can't open the channel
	}

	go handleOutbound(*outbound, rdb, dg) // Start listening for outbound messages

	// Wait here until CTRL-C or other term signal is received.
	log.Info().Str("func", "main").Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)                                             // Create a channel to listen for signals
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill) // Listen for signals
	<-sc
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	str, _ := json.Marshal(m)         // Convert the message to JSON so that we can send it to Redis
	msg := &model.Message{}           // Create a new message
	err := msg.Unmarshal([]byte(str)) // Unmarshal the JSON into the message
	if err != nil {
		log.Err(err).Str("func", "messageCreate").Str("id", msg.ID).Msg("Unable to unmarshal message")
		return
	}

	msg.Metadata.ID = uuid.Must(uuid.NewV4(), *new(error)) // Generate a UUID for the message
	msg.Metadata.Source = *id                              // Set the source to the ID of this gateway
	stringMsg, _ := json.Marshal(msg)                      // Convert the message to JSON so that we can send it to Redis
	rdb.Publish(ctx, *inbound, stringMsg)                  // Publish the message to Redis
	log.Debug().Str("func", "messageCreate").Str("id", msg.Metadata.ID.String()).Msg("Published message to " + *inbound)
}

// handleOutbound handles outbound messages from Redis destined for Discord
// It will subscribe to the given topic and send any messages it receives to Discord
// Unforunately it uses a slightly different message format than the inbound messages
// See model.MessageSend for more information
func handleOutbound(sub string, rdb *redis.Client, s *discordgo.Session) {
	log.Info().Str("func", "handleOutbound").Msg("Listening for outbound messages on topic '" + sub + "'")
	ctx := context.Background()      // Redis context
	topic := rdb.Subscribe(ctx, sub) // Subscribe to the given topic
	channel := topic.Channel()       // Setup a Go channel to capture messages from Redis

	for msg := range channel { // Loop through the messages as they come in
		log.Debug().Str("func", "handleOutbound").Msg("Message received")
		m := &model.MessageSend{}               // Create a new message
		err := m.Unmarshal([]byte(msg.Payload)) // Unpack the message from Redis into Payload field
		if err != nil {
			log.Err(err).Str("func", "handleOutbound").Msg("Unable to unmarshal message")
		}

		log.Debug().
			Str("func", "handleOutbound").
			Str("id", m.Metadata.ID.String()).
			Bool("will_send", m.Metadata.Dest == *id).
			Str("channel", m.ChannelID).
			Str("content", m.Content)
		if m.Metadata.Dest == *id { // Check if the message is for this gateway, if not, ignore it
			s.ChannelMessageSend(m.ChannelID, m.Content)
		}
	}
}

// rdbConnect connects to Redis and returns a client
func rdbConnect(addr, pass string, db int) *redis.Client {
	log.Debug().Str("func", "rdbConnect").
		Str("addr", addr).
		Int("db", db).
		Msg("Connecting to Redis")

	ctx := context.Background() // Redis context
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr, // Redis address
		Password: pass, // Redis password
		DB:       0,    // use default DB
	})

	err := rdb.Ping(ctx).Err() // Ping Redis to make sure we can connect
	if err != nil {
		// if we can't connect, try again in 3 seconds then succeed or give up
		// yes this is lazy
		log.Err(err).
			Str("func", "rdbConnect").
			Msg("Unable to connect to Redis, trying again in 3 seconds")
		time.Sleep(3 * time.Second)
		err := rdb.Ping(ctx).Err()
		if err != nil {
			log.Err(err).
				Str("func", "rdbConnect").
				Msg("Unable to connect to Redis, exiting")
			os.Exit(1)
		}
	}

	return rdb
}
