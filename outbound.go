package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

// handleOutbound handles outbound messages from Redis destined for Discord
// It will subscribe to the given topic and send any messages it receives to Discord
// If a reply or mention is requested, it will attempt to handle that as well
// Unforunately it uses a slightly different message format than the inbound messages
// See model.MessageSend for more information
func handleOutbound(sub string, rdb *redis.Client, s *discordgo.Session, ctx context.Context) {
	// Subscribe to the given topic
	pubsub := rdb.Subscribe(redisCtx, sub)
	// Get the channel for the subscription
	outbound := pubsub.Channel()

	// Loop forever
	for {
		// Wait for a message
		msg := <-outbound

		// Do this before unmarshalling the message
		// Unmarshaling can crash the process and we want to log the message before that happens
		log.Debug().
			Str("topic", sub).
			Str("payload", msg.Payload).
			Msg("Received message from Redis")
	}
}
