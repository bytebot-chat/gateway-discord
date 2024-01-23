package main

import (
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
	// Example for sithmail: discord.sithmail.#channel.user
	topic := "discord." + m.GuildID + "." + m.ChannelID + "." + m.Author.ID

	// Publish the message to redis
	res := rdb.Publish(redisCtx, topic, m)
	if res.Err() != nil {
		log.Err(res.Err()).Msg("Unable to publish message")
		return
	}

	log.Info().
		Str("func", "messageCreate").
		Msgf("Published message to %s", topic)

	log.Debug().
		Str("func", "messageCreate").
		Msg("Sent message to Redis")
}
