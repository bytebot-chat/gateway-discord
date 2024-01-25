package main

import (
	"context"
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

// handleOutbound handles outbound messages from Redis destined for Discord
// It will subscribe to the given topic and send any messages it receives to Discord
// If a reply or mention is requested, it will attempt to handle that as well
// Unforunately it uses a slightly different message format than the inbound messages
// See model.MessageSend for more information
func handleOutbound(ctx context.Context, rdb *redis.Client, s *discordgo.Session) {
	// Subscribe to the given topic
	pubsub := rdb.PSubscribe(ctx, "outbound.discord.*")
	defer pubsub.Close()

	log.Info().
		Str("topic", "outbound.discord.*").
		Msg("Subscribed to Redis topic")

	// Get the channel for the subscription
	outbound := pubsub.Channel()

	// Loop forever
	for {
		// Wait for a message
		msg := <-outbound

		// Do this before unmarshalling the message
		// Unmarshaling can crash the process and we want to log the message before that happens
		log.Debug().
			Str("topic", msg.Channel).
			Str("payload", msg.Payload).
			Msg("Received message from Redis")

		// Unmarhal the message and construct a discordgo.MessageSend object
		var outboundMsg string
		err := json.Unmarshal([]byte(msg.Payload), &outboundMsg)
		if err != nil {
			log.Err(err).
				Str("topic", msg.Channel).
				Str("payload", msg.Payload).
				Msg("Unable to unmarshal message from Redis")
			continue
		}

		topic, err := newPubsubDiscordTopicAddr(msg.Channel)
		if err != nil {
			log.Err(err).
				Str("topic", msg.Channel).
				Msg("Unable to parse topic")
			continue
		}
		s.ChannelMessageSend(topic.ChannelID, outboundMsg)
	}

}
