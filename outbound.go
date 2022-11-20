package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/bytebot-chat/gateway-discord/model"
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

		// Unmarshal the message into a MessageSend
		m := &model.MessageSend{}
		err := m.UnmarshalJSON([]byte(msg.Payload))

		if err != nil {
			log.Err(err).
				Str("func", "handleOutbound").
				Str("topic", sub).
				Msg("Unable to unmarshal message")
			continue
		}

		// Send the message to Discord
		if m.ShouldReply {
			log.Debug().
				Str("func", "handleOutbound").
				Str("id", m.Metadata.ID.String()).
				Msg("Reply requested")

			_, _ = s.ChannelMessageSendReply(m.ChannelID, m.Content, m.PreviousMessage.Reference())
		} else {
			_, err = s.ChannelMessageSend(m.ChannelID, m.Content) // TODO: Handle replies and mentions
			if err != nil {
				log.Err(err).
					Str("func", "handleOutbound").
					Str("id", m.Metadata.ID.String()).
					Str("topic", sub).
					Msg("Unable to send message to Discord")
				continue
			}
		}

		log.Debug().
			Str("func", "handleOutbound").
			Str("id", m.Metadata.ID.String()).
			Str("source", m.Metadata.Source).
			Str("dest", m.Metadata.Dest).
			Str("topic", sub).
			Msg("Sent message to Discord")
	}
}
