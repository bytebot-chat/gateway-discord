package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bbriggs/bytebot-discord/model"
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis/v8"
	"github.com/satori/go.uuid"
)

// Variables used for command line parameters
var (
	Token string
	ctx   context.Context
	rdb   *redis.Client

	serv      = flag.String("server", "localhost:6667", "hostname and port for irc server to connect to")
	redisAddr = flag.String("redis", "localhost:6379", "Address and port of redis host")
	nick      = flag.String("nick", "bytebot", "nickname for the bot")
	id        = flag.String("id", "irc", "ID to use when publishing messages")
	inbound   = flag.String("inbound", "irc-inbound", "Pubsub queue to publish inbound messages to")
	outbound  = flag.String("outbound", *id, "Pubsub to subscribe to for sending outbound messages. Defaults to being equivalent to `id`")
	tls       = flag.Bool("tls", false, "Use TLS when connecting to IRC server")
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	rdb = rdbConnect(*redisAddr)
	ctx = context.Background()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

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

	return
}

func handleOutbound(sub string, rdb *redis.Client, s *discordgo.Session) {
	ctx := context.Background()
	topic := rdb.Subscribe(ctx, sub)
	channel := topic.Channel()
	for msg := range channel {
		m := &model.Message{}
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

	return rdb
}
