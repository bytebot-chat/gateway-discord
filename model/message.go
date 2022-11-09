package model

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	uuid "github.com/satori/go.uuid"
)

// Message is the struct that is used to pass messages from the Gateway to the Redis pubsub (inbound messages)
type Message struct {
	*discordgo.Message
	Metadata Metadata `json:"metadata"`
}

// MessageSend is the struct that is used to pass messages from the Redis pubsub to the Discord Gateway (outbound messages)
// Because the discordgo.Session.ChannelMessageSend() method only accepts channel ID and content as a string, our struct is limited to those two fields as well.
// Future work may expand this to include more fields or expand metadata to include more information that can be used to forumlate more complex responses.
type MessageSend struct {
	ChannelID string   `json:"channel_id"` // ChannelID is the ID of the discord channel to send the message to
	Content   string   `json:"content"`    // Content is the text body of the message to send
	Metadata  Metadata `json:"metadata"`
}

// Metadata is used by the Gateway(s) and app(s) to trace messages and identify intended recipients
type Metadata struct {
	Source string    // Source is the ID of the Gateway or App that sent the message
	Dest   string    // Dest is the ID of the Gateway or App that the message is intended for
	ID     uuid.UUID // ID is a UUID that is generated for each message
}

// Marhsal converts the message to JSON
func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// MarshalReply converts the message to JSON and adds the metadata from the original message
// MarshalReply sends a response to the originating channel or direct message but does not do a "discord reply"
func (m *Message) MarshalReply(meta Metadata, dest string, s string) ([]byte, error) {
	reply := &MessageSend{
		Content:   s,
		ChannelID: dest,
		Metadata:  meta,
	}
	return json.Marshal(reply)
}

// Unmarshal converts the JSON (in bytes) to a message
// Example:
// 	msg := &model.Message{}
// 	if err := msg.Unmarshal([]byte(`{"content":"hello world"}`)); err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println(msg.Content)
func (m *Message) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, m); err != nil {
		return err
	}
	return nil
}

// UnmarshalReply converts the JSON (in bytes) to a message
// Example:
// 	msg := &model.MessageSend{}
// 	if err := msg.Unmarshal([]byte(`{"content":"hello world"}`)); err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println(msg.Content)
func (m *MessageSend) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, m); err != nil {
		return err
	}
	return nil
}
