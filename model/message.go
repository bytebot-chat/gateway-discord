package model

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	"github.com/satori/go.uuid"
)

type Message struct {
	*discordgo.Message
	Metadata Metadata
}

type MessageSend struct {
	ChannelID string
	Content   string
	Metadata  Metadata
}

type Metadata struct {
	Source string
	Dest   string
	ID     uuid.UUID
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) MarshalReply(meta Metadata, dest string, s string) ([]byte, error) {
	reply := &MessageSend{
		Content:   s,
		ChannelID: dest,
		Metadata:  meta,
	}
	return json.Marshal(reply)
}

func (m *Message) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, m); err != nil {
		return err
	}
	return nil
}

func (m *MessageSend) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, m); err != nil {
		return err
	}
	return nil
}
