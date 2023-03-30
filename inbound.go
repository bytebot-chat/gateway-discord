package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bytebot-chat/gateway-discord/model"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

// This function is called every time a new message is created from any Message
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	log.Debug().
		Str("func", "messageCreate").
		Str("id", m.ID).
		Str("source", m.Author.ID+":"+m.Author.Discriminator).
		Msg("Received message from Discord")

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Create a new message and populate it with the data from the Discord message
	msg := &model.Message{
		Message: m.Message,
	}

	// Set the metadata before sending it to Redis
	msg.Metadata = model.Metadata{
		Source: *id,
		Dest:   "",
		ID:     uuid.NewV4(),
	}

	// Marshal the message to JSON
	json, err := msg.MarshalJSON()
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal message")
		return
	}

	// Publish the message to redis
	res := rdb.Publish(redisCtx, *inbound, json)
	if res.Err() != nil {
		log.Err(res.Err()).Msg("Unable to publish message")
		return
	}

	log.Info().
		Str("func", "messageCreate").
		Str("id", msg.Metadata.ID.String()).
		Msgf("Published message to %s", *inbound)

	log.Debug().
		Str("func", "messageCreate").
		Str("id", msg.Metadata.ID.String()).
		Str("source", msg.Metadata.Source).
		Str("dest", msg.Metadata.Dest).
		Str("topic", *inbound).
		Str("content", msg.Content).
		Str("author", msg.Author.ID+":"+msg.Author.Discriminator).
		Msg("Sent message to Redis")
}
