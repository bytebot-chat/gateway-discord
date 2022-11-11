package model

import (
	"encoding/json"

	uuid "github.com/satori/go.uuid"
)

// MessageSend is the struct that is used to pass messages from the Redis pubsub to the Discord Gateway (outbound messages)
// Because the discordgo.Session.ChannelMessageSend() method only accepts channel ID and content as a string, our struct limits iteslef to those two fields as well.
// Future work may expand this to include more fields or expand metadata to include more information that can be used to forumlate more complex responses.
type MessageSend struct {
	ChannelID string   `json:"channel_id,omitempty"` // ChannelID is the ID of the discord channel to send the message to
	Content   string   `json:"content,omitempty"`    // Content is the text body of the message to send
	Metadata  Metadata `json:"metadata,omitempty"`
}

// Deprecated in favor of newer methods that consume the entire model.Message struct
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

// RespondToChannelOrThread generates a MessageSend struct that can be used to respond to a channel or thread
// It optionally allows the message to reply or mention the user that sent the original message
func (m *Message) RespondToChannelOrThread(sourceApp, content string, shouldReply, shouldMention bool) *MessageSend {

	return &MessageSend{
		Content:   content,     // Actual text to send
		ChannelID: m.ChannelID, // Send the message to the channel or thread that the original message was sent from
		Metadata: Metadata{
			Source: sourceApp,         // the ID of the app sending the message
			Dest:   m.Metadata.Source, // The destination of the message is the ID of the app that is receiving the message (ie, 'discord')
			ID:     uuid.NewV4(),      // Generate a new UUID for the message
		},
	}
}
