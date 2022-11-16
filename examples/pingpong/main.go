package main

import (
	"context"
	"flag"
	"os"
	"strings"
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

func init() {
	flag.Parse()
}

func main() {

	// An example of logging using zerolog
	log.Info().
		Str("version", "0.0.1").
		Str("addr", *addr).
		Str("inbound", *inbound).
		Str("outbound", *outbound).
		Msg("Starting pingpong")

	// Create a new Redis client
	ctx := context.Background() // Redis context used for all Redis operations
	log.Info().
		Str("address", *addr).
		Msg("Connecting to redis")

	rdb := redis.NewClient(&redis.Options{
		Addr: *addr,
		DB:   0, // use default DB
	})

	err := rdb.Ping(ctx).Err() // Ping redis to make sure it's up

	// If there is an error, log it and exit
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

	// Messages come in from a pubsub queue
	// We need to subscribe to the queue and listen for messages
	// We use topic.Channel() to get the channel to listen on for messages and then use a for loop to listen for messages
	// When a message comes in, we need to parse it and then send a response
	// We use the MessageSend struct to send the response
	// We then use the rdb.Publish() method to send the response to the outbound queue

	// Subscribe to the inbound queue
	topic := rdb.Subscribe(ctx, *inbound)
	channel := topic.Channel()

	// Listen for messages
	for msg := range channel {
		log.Debug().
			Msg("Received message")

		// Create a new MessageSend struct
		var message model.Message

		// Unmarshal the message bytes into the struct
		err := message.UnmarshalJSON([]byte(msg.Payload))

		// If there is an error, log it and continue
		if err != nil {
			log.Err(err).
				Msg("Unable to parse message")

			// If we can't parse the message, we can't send a response
			// So we just continue to the next message
			continue
		}

		// Check if the message is a ping
		if !strings.HasPrefix(message.Content, "ping") {
			// If it's not a ping, we don't need to respond
			continue
		}

		// Log that we're sending a response
		log.Info().
			Msg("Ping received. Sending pong")

		// Use a convenience method to create a new MessageSend struct
		// This method takes the app name, the content of the message to send, whether to reply to the message, and whether to mention the user who sent the message
		resp := message.RespondToChannelOrThread(APP_NAME, "Pong with reply", true, false)

		// Debug log the response
		log.Debug().
			Str("message", resp.Content).
			Str("dest", resp.Metadata.Dest).
			Msg("Sending message")

		// Marshal the struct into bytes
		respBytes, err := resp.MarshalJSON()

		// If there is an error, log it and continue
		if err != nil {
			log.Err(err).
				Str("message", msg.Payload).
				Msg("Unable to marshal message")
			continue
		}

		// Publish the message to the outbound queue
		err = rdb.Publish(ctx, *outbound, respBytes).Err()

		// If there is an error, log it and continue
		if err != nil {
			log.Err(err).
				Str("message", msg.Payload).
				Msg("Unable to publish message")
			continue
		}
	}

}
