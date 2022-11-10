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
	log.Debug().Msg("Starting up!")
	flag.Parse()
	ctx := context.Background() // Redis context
	log.Debug().Str("address", *addr).Msg("Connecting to redis")
	rdb := redis.NewClient(&redis.Options{ // Connect to redis
		Addr: *addr, // Redis address from command line
		DB:   0,     // Use default DB
	})

	err := rdb.Ping(ctx).Err() // Ping redis to make sure it's up
	if err != nil {
		log.Err(err).Msg("Unable to continue without connection. Trying again in 3 seconds")
		time.Sleep(3 * time.Second) // Wait 3 seconds before trying again
		err := rdb.Ping(ctx).Err()
		if err != nil {
			log.Err(err).Msg("Unable to continue without connection. Exiting!")
			os.Exit(1)
		}
	}

	log.Debug().Str("topic", *inbound).Msg("Subscribing to inbound topic")
	topic := rdb.Subscribe(ctx, *inbound) // Subscribe to the inbound topic
	channel := topic.Channel()            // Create a channel to listen for messages on
	log.Debug().Msg("Listening for messages...")
	for msg := range channel { // Read messages from the channel in a loop
		m := &model.Message{}                   // Create a new message
		err := m.Unmarshal([]byte(msg.Payload)) // Unmarshal the message
		if err != nil {
			log.Err(err).Msg("Unable to unmarshal message")
		}
		if m.Content == "ping" {
			log.Debug().Str("id", m.ID).Msg("Received ping")
			reply(ctx, *m, rdb) // Reply to the message
		}
	}
}

// reply replies to a message
func reply(ctx context.Context, m model.Message, rdb *redis.Client) {
	log.Debug().Msg("Replying to message")
	stringMsg, _ := m.RespondToChannelOrThread("discord-pingpong", "pong") // Marshal the content into a reply
	rdb.Publish(ctx, *outbound, stringMsg)                                 // Publish the message to the outbound topic
	log.Debug().Str("topic", *outbound).Msg("Published message")
}
