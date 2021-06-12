package model

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	"github.com/satori/go.uuid"
)

type Message struct {
	*discordgo.MessageCreate
	Metadata Metadata
}

type Metadata struct {
	Source string
	Dest   string
	ID     uuid.UUID
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, m); err != nil {
		return err
	}
	return nil
}
