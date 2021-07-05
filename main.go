package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/bytebot-chat/gateway-discord/model"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/satori/go.uuid"
)

// Variables used for command line parameters
var (
	Token string
	ctx   context.Context
	rdb   *redis.Client

	redisAddr = flag.String("redis", "localhost:6379", "Address and port of redis host")
	id        = flag.String("id", "discord", "ID to use when publishing messages")
	inbound   = flag.String("inbound", "discord-inbound", "Pubsub queue to publish inbound messages to")
	outbound  = flag.String("outbound", *id, "Pubsub to subscribe to for sending outbound messages. Defaults to being equivalent to `id`")
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	log.Info().
		Str("Redis address", *redisAddr).
		Msg("Discord starting up!")

	rdb = rdbConnect(*redisAddr)
	ctx = context.Background()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Err(err).Msg("Unable to continue without connection. Exiting!")
		os.Exit(1)
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	fmt.Println("Publishing inbound messages on topic '" + *inbound + "'")
	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		log.Err(err).Msg("Unable to open channel for reading messages")
		os.Exit(1)
	}

	go handleOutbound(*id, rdb, dg)
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	str, _ := json.Marshal(m)
	log.Debug().
		RawJSON("Received message", []byte(str)).
		Msg("Received message")

	msg := &model.Message{}
	err := msg.Unmarshal([]byte(str))
	if err != nil {
		fmt.Println(err)
		return
	}
	msg.Metadata.ID = uuid.Must(uuid.NewV4(), *new(error))
	msg.Metadata.Source = *id
	stringMsg, _ := json.Marshal(msg)
	rdb.Publish(ctx, *inbound, stringMsg)
	log.Debug().Msg("Published message to " + *inbound)
	return
}

func handleOutbound(sub string, rdb *redis.Client, s *discordgo.Session) {
	log.Info().Msg("Listening for outbound messages on topic '" + sub + "'")
	ctx := context.Background()
	topic := rdb.Subscribe(ctx, sub)
	channel := topic.Channel()
	for msg := range channel {
		m := &model.MessageSend{}
		err := m.Unmarshal([]byte(msg.Payload))
		if err != nil {
			fmt.Println(err)
		}
		if m.Metadata.Dest == *id {
			s.ChannelMessageSend(m.ChannelID, m.Content)
		}
	}
}

func rdbConnect(addr string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		time.Sleep(3 * time.Second)
		err := rdb.Ping(ctx).Err()
		if err != nil {
			log.Crit("FATAL unable to connect to redis", "error", err)
			os.Exit(1)
		}
	}

	return rdb
}
