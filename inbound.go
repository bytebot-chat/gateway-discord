package main

import (
	"context"
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	"github.com/bytebot-chat/gateway-discord/model"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	str, _ := json.Marshal(m)         // Convert the message to JSON so that we can send it to Redis
	msg := &model.Message{}           // Create a new message
	err := msg.Unmarshal([]byte(str)) // Unmarshal the JSON into the message
	if err != nil {
		log.Err(err).
			Str("func", "messageCreate").
			Str("id", msg.ID).
			Msg("Unable to unmarshal message")
		return
	}

	msg.Metadata.ID = uuid.Must(uuid.NewV4(), *new(error)) // Generate a UUID for the message
	msg.Metadata.Source = *id                              // Set the source to the ID of this gateway
	stringMsg, _ := json.Marshal(msg)                      // Convert the message to JSON so that we can send it to Redis
	res := rdb.Publish(ctx, *inbound, stringMsg)           // Publish the message to Redis
	if res.Err() != nil {
		log.Err(res.Err()).
			Str("func", "messageCreate").
			Str("id", msg.ID).
			Msg("Unable to publish message to Redis")
		return
	}

	log.Debug().
		Str("func", "messageCreate").
		Str("id", msg.Metadata.ID.String()).
		Str("source", msg.Metadata.Source).
		Str("dest", msg.Metadata.Dest).
		Str("topic", *inbound).
		Msg("Published message to Redis")
}

// handleInbound handles inbound messages from Discord destined for Redis
// It will subscribe to the given topic and send any messages it receives to Redis
// See model.Message for more information
func handleInbound(sub string, rdb *redis.Client, s *discordgo.Session) {
	log.Info().
		Str("func", "handleInbound").
		Msg("Listening for inbound messages on topic '" + sub + "'")

	ctx := context.Background()      // Redis context
	topic := rdb.Subscribe(ctx, sub) // Subscribe to the given topic
	channel := topic.Channel()       // Setup a Go channel to capture messages from Redis

	for msg := range channel { // Loop through the messages as they come in
		log.Debug().
			Str("func", "handleInbound").
			Str("topic", sub).
			Msg("Received message from Redis")

		// Create a new message and unmarshal the JSON into it
		message := &model.Message{}
		err := message.UnmarshalJSON([]byte(msg.Payload))
		if err != nil {
			log.Err(err).
				Str("func", "handleInbound").
				Str("topic", sub).
				Msg("Unable to unmarshal message")
			continue
		}

		// Publish the message to redis
		res := rdb.Publish(ctx, *inbound, msg.Payload)
		if res.Err() != nil {
			log.Err(res.Err()).
				Str("func", "handleInbound").
				Str("topic", sub).
				Msg("Unable to publish message to Redis")
			continue
		}
	}
}
