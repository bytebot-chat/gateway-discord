package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bytebot-chat/gateway-discord/model"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

const APP_NAME = "pingpong"

var (
	addr     = flag.String("redis", "localhost:6379", "Redis server address")
	inbound  = flag.String("inbound", "discord-inbound", "Pubsub queue to listen for new messages")
	outbound = flag.String("outbound", "discord-outbound", "Pubsub queue for sending messages outbound")
)

func main() {
	flag.Parse()
	log.Info().
		Str("version", "0.0.1").
		Str("addr", *addr).
		Str("inbound", *inbound).
		Str("outbound", *outbound).
		Msg("Starting pingpong")

	ctx := context.Background() // Redis context
	log.Debug().
		Str("address", *addr).
		Msg("Connecting to redis")
	rdb := redis.NewClient(&redis.Options{ // Connect to redis
		Addr: *addr, // Redis address from command line
		DB:   0,     // Use default DB
	})

	err := rdb.Ping(ctx).Err() // Ping redis to make sure it's up
	if err != nil {
		log.Err(err).
			Str("address", *addr).
			Msg("Unable to continue without connection. Trying again in 3 seconds")
		time.Sleep(3 * time.Second) // Wait 3 seconds before trying again
		err := rdb.Ping(ctx).Err()
		if err != nil {
			log.Err(err).
				Str("address", *addr).
				Msg("Unable to continue without connection. Exiting!")
			os.Exit(1)
		}
	}

	topic := rdb.Subscribe(ctx, *inbound) // Subscribe to the inbound topic
	channel := topic.Channel()            // Create a channel to listen for messages on
	log.Debug().
		Str("topic", *inbound).
		Msg("Subscribed to topic")
	for msg := range channel { // Read messages from the channel in a loop
		fmt.Println(msg.Payload)                    // Print the message payload
		m := &model.Message{}                       // Create a new message
		err := m.UnmarshalJSON([]byte(msg.Payload)) // Unmarshal the message
		if err != nil {
			log.Err(err).
				Str("topic", *inbound).
				Msg("Unable to unmarshal message")
			continue
		}
		if m.Content == "ping" {
			log.Debug().
				Str("id", m.ID).
				Str("source", m.Metadata.Source).
				Str("dest", m.Metadata.Dest).
				Str("channel", m.ChannelID).
				Msg("Received ping")
			reply(ctx, *m, rdb) // Reply to the message
		}
	}
}

// reply replies to a message
func reply(ctx context.Context, m model.Message, rdb *redis.Client) {
	// Create a new message
	// Respond in same channel, do not reply or mention the user
	msg := m.RespondToChannelOrThread(APP_NAME, "pong", false, false)
	msg.Content = m.Content // Set the content to the original message

	// Marshal the message
	b, err := msg.MarshalJSON()
	if err != nil {
		log.Err(err).
			Str("id", m.ID).
			Str("source", m.Metadata.Source).
			Str("dest", m.Metadata.Dest).
			Str("channel", m.ChannelID).
			Msg("Unable to marshal message")
	}

	// Publish the message to the outbound topic
	err = rdb.Publish(ctx, *outbound, b).Err()
	if err != nil {
		log.Err(err).
			Str("id", m.ID).
			Str("source", m.Metadata.Source).
			Str("dest", m.Metadata.Dest).
			Str("channel", m.ChannelID).
			Msg("Unable to publish message")
	}
}
