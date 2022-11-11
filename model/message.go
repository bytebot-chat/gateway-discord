package model

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	uuid "github.com/satori/go.uuid"
)

// Message is the struct that is used to pass messages from the Gateway to the Redis pubsub (inbound messages)
type Message struct {
	*discordgo.Message `json:",omitempty"`
	Metadata           Metadata `json:"metadata"`
}

// MessageSend is the struct that is used to pass messages from the Redis pubsub to the Discord Gateway (outbound messages)
// Because the discordgo.Session.ChannelMessageSend() method only accepts channel ID and content as a string, our struct limits iteslef to those two fields as well.
// Future work may expand this to include more fields or expand metadata to include more information that can be used to forumlate more complex responses.
type MessageSend struct {
	ChannelID string   `json:"channel_id,omitempty"` // ChannelID is the ID of the discord channel to send the message to
	Content   string   `json:"content,omitempty"`    // Content is the text body of the message to send
	Metadata  Metadata `json:"metadata,omitempty"`
}

// Metadata is used by the Gateway(s) and app(s) to trace messages and identify intended recipients
type Metadata struct {
	Source string    `json:"source,omitempty"` // Source is the ID of the Gateway or App that sent the message
	Dest   string    `json:"dest,omitempty"`   // Dest is the ID of the Gateway or App that the message is intended for
	ID     uuid.UUID `json:"id,omitempty"`     // ID is a UUID that is generated for each message
}

// Marhsal converts the message to JSON
func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal converts the JSON (in bytes) to a message
// This method is deprecate in favor of the UnmarshalJSON method and will be removed in a future release
// Correct behavior from this method is not guaranteed
// Example:
// 	msg := &model.Message{}
// 	if err := msg.Unmarshal([]byte(`{"content":"hello world"}`)); err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println(msg.Content)
func (m *Message) Unmarshal(b []byte) error {
	var byteMsg []byte
	if err := json.Unmarshal(b, &byteMsg); err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON converts the JSON (in bytes) to a message
// Because the *discordgo.Message struct is embedded in the Message struct and also has a MarshalJSON method,
// go will call the MarshalJSON method of the *discordgo.Message struct when the Message struct is marshaled
// unless we override it with a custom MarshalJSON method in the Message struct, which we do
// Example:
// 	msg := &model.Message{}
// 	if err := msg.UnmarshalJSON([]byte(`{"content":"hello world"}`)); err != nil {
// 		log.Fatal(err)
// 	}
func (m *Message) UnmarshalJSON(b []byte) error {
	msg := make(map[string]json.RawMessage)

	if err := json.Unmarshal(b, &msg); err != nil {
		return err
	}

	if err := json.Unmarshal(msg["message"], &m.Message); err != nil {
		return err
	}

	if err := json.Unmarshal(msg["metadata"], &m.Metadata); err != nil {
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
