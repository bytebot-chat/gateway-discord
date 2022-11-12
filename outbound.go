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
// Unforunately it uses a slightly different message format than the inbound messages
// See model.MessageSend for more information
func handleOutbound(sub string, rdb *redis.Client, s *discordgo.Session) {
	log.Info().Str("func", "handleOutbound").Msg("Listening for outbound messages on topic '" + sub + "'")
	ctx := context.Background()      // Redis context
	topic := rdb.Subscribe(ctx, sub) // Subscribe to the given topic
	channel := topic.Channel()       // Setup a Go channel to capture messages from Redis

	for msg := range channel { // Loop through the messages as they come in
		log.Debug().Str("func", "handleOutbound").
			Str("topic", sub).
			Str("payload", msg.Payload).
			Send()

		m := &model.MessageSend{}               // Create a new message
		err := m.Unmarshal([]byte(msg.Payload)) // Unpack the message from Redis into Payload field
		if err != nil {
			log.Err(err).Str("func", "handleOutbound").Msg("Unable to unmarshal message") // TODO: Fix this log line
		}

		log.Debug().
			Str("func", "handleOutbound").
			Str("message_id", m.Metadata.ID.String()).
			Str("message_source", m.Metadata.Source).
			Str("message_destination", m.Metadata.Dest).
			Bool("will_send", m.Metadata.Dest == *id).
			Str("channel", m.ChannelID).
			Str("content", m.Content).
			Send()

		if m.Metadata.Dest == *id { // Check if the message is for this gateway, if not, ignore it
			_, err := s.ChannelMessageSend(m.ChannelID, m.Content) // Send the message to Discord
			if err != nil {
				log.Err(err).
					Str("func", "handleOutbound").
					Msg("Unable to send message to Discord")
			}
		}
	}
}
