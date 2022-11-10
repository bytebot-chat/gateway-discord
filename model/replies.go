package model

import (
	"encoding/json"

	uuid "github.com/satori/go.uuid"
)

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

// RespondToChannelOrThread sends a message to the channel or thread that the original message was sent from
// TODO: Make this actually work. It's just stubbed out for now.
func (m *Message) RespondToChannelOrThread(sourceApp, content string) ([]byte, error) {

	reply := &MessageSend{
		Content:   content,     // Actual text to send
		ChannelID: m.ChannelID, // Send the message to the channel or thread that the original message was sent from
		Metadata: Metadata{
			Source: sourceApp,         // the ID of the app sending the message
			Dest:   m.Metadata.Source, // The destination of the message is the ID of the app that is receiving the message (ie, 'discord')
			ID:     uuid.NewV4(),      // Generate a new UUID for the message
		},
	}
	return json.Marshal(reply)
}

// ReplyToUser sends a message to the user that the original message was sent from by using a "discord reply"
// TODO: Make this actually work. It's just stubbed out for now.
func (m *Message) ReplyToUser(meta Metadata, dest string, s string) ([]byte, error) {
	reply := &MessageSend{
		Content:   s,
		ChannelID: dest,
		Metadata:  meta,
	}
	return json.Marshal(reply)
}

// DMReplyToUser responds to the user via Direct Message
// TODO: Make this actually work. It's just stubbed out for now.
func (m *Message) DMReplyToUser(meta Metadata, dest string, s string) ([]byte, error) {
	reply := &MessageSend{
		Content:   s,
		ChannelID: dest,
		Metadata:  meta,
	}
	return json.Marshal(reply)
}
