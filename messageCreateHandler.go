package main

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
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

	// Topic ID is derived from the protocol, server, channel, and user and dot-delimited
	// Example for sithmail: inbound.discord.sithmail.#channel.user
	topic := "inbound.discord." + m.GuildID + "." + m.ChannelID + "." + m.Author.ID

	// Marshal the message into a byte array
	// This is required because Redis only accepts byte arrays
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		log.Err(err).Msg("Failed to marshal message into JSON")
		return
	}

	// Publish the message to redis
	res := rdb.Publish(redisCtx, topic, jsonBytes)
	if res.Err() != nil {
		log.Err(res.Err()).Msg("Unable to publish message")
		return
	}

	log.Info().
		Str("func", "messageCreate").
		Msgf("Published message to %s", topic)

	log.Debug().
		Str("func", "messageCreate").
		Str("message_id", m.ID).
		Str("source_username", m.Author.Username).
		Str("dest_topic", topic).
		Str("message_content", m.Content).
		Msg("Sent message to Redis")
}
