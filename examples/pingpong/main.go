package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/bytebot-chat/gateway-discord/model"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

var addr = flag.String("redis", "localhost:6379", "Redis server address")
var inbound = flag.String("inbound", "discord-inbound", "Pubsub queue to listen for new messages")
var outbound = flag.String("outbound", "discord-outbound", "Pubsub queue for sending messages outbound")

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
		m := &model.Message{}                   // Create a new message
		err := m.Unmarshal([]byte(msg.Payload)) // Unmarshal the message
		if err != nil {
			log.Err(err).
				Str("topic", *inbound).
				Msg("Unable to unmarshal message")
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
	log.Debug().
		Str("id", m.ID).
		Str("source", m.Metadata.Source).
		Str("dest", m.Metadata.Dest).
		Str("channel", m.ChannelID).
		Msg("Replying to message")
	stringMsg, _ := m.RespondToChannelOrThread("discord-pingpong", "pong") // Marshal the content into a reply
	rdb.Publish(ctx, *outbound, stringMsg)                                 // Publish the message to the outbound topic
	log.Debug().
		Str("id", m.ID).
		Str("source", m.Metadata.Source).
		Str("dest", m.Metadata.Dest).
		Str("topic", *outbound).
		Msg("Published message")
}
