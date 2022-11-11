package model

import (
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/r3labs/diff/v3"
	uuid "github.com/satori/go.uuid"
)

func TestMessage_RespondToChannelOrThread(t *testing.T) {
	/*
		This test case checks that the RespondToChannelOrThread method returns the correct JSON
		The function is intended to send a message to the channel or thread that the original message was sent from
		This means the ChannelID of the message should be the same as the ChannelID of the original message and
		the Metadata.Dest should be the same as the Metadata.Source of the original message
	*/

	// Setup test cases and expected results through this struct
	// The test cases are the values of the Message struct that is passed to the RespondToChannelOrThread method
	// The expected results are the values that the method should return once it has been
	// marshaled into JSON and unmarshaled back into a MessageSend struct
	tests := []struct {
		name          string       // name of the test
		message       *Message     // Original message to respond to
		want          *MessageSend // Expected response
		sourceApp     string       // ID of the app sending the response
		content       string       // Text to send in the response
		shouldReply   bool         // Whether the response should be a reply to the original message
		shouldMention bool         // Whether the response should mention the original message author
		wantErr       bool         // Whether or not the test should fail
	}{
		{
			name: "Basic, no replies",
			message: &Message{
				Message: &discordgo.Message{
					Content:   TestInboundMessageBody,
					ChannelID: TestChannelID,
				},
				Metadata: Metadata{
					Source: "gateway",
					Dest:   "",
				},
			},
			want: &MessageSend{
				Content:   TestOutboundMessageBody,
				ChannelID: TestChannelID,
				Metadata: Metadata{
					Source: TestAppName,
					Dest:   TestMetdataSource,
					ID:     uuid.FromStringOrNil(TestOutboundMetadataUUID),
				},
			},
			sourceApp:     TestAppName,
			content:       TestOutboundMessageBody,
			wantErr:       false,
			shouldReply:   false,
			shouldMention: false,
		},
		{
			name: "Reply with no mention",
			message: &Message{
				Message: &discordgo.Message{
					ID:        TestInboundDiscordMessageID,
					ChannelID: TestChannelID,
					Content:   TestInboundMessageBody,
				},
				Metadata: Metadata{
					Source:      "gateway",
					Dest:        "",
					ID:          uuid.FromStringOrNil(TestInboundMetadataUUID),
					Reply:       false, // Inbound message is not a reply
					InReplyTo:   "",    // Inbound message is not a reply
					MentionUser: false, // Inbound message does not mention the user
				},
			},
			want: &MessageSend{
				Content:   TestOutboundMessageBody,
				ChannelID: TestChannelID, // ChannelID should be the same as the original message
				Metadata: Metadata{
					Source:      TestAppName,                                    // Source should be the app sending the response
					Dest:        TestMetdataSource,                              // Dest should be the source from the original message
					ID:          uuid.FromStringOrNil(TestOutboundMetadataUUID), // ID should be a new UUID
					Reply:       true,                                           // Outbound message should be a reply to the original message
					InReplyTo:   TestInboundDiscordMessageID,                    // Outbound message should be a reply to the original message
					MentionUser: false,                                          // Outbound message should not mention the user
				},
			},
			sourceApp:     TestAppName,
			content:       TestOutboundMessageBody,
			wantErr:       false,
			shouldReply:   true,
			shouldMention: false,
		},
		{
			name: "Reply with mention",
			message: &Message{
				Message: &discordgo.Message{
					ID:        TestInboundDiscordMessageID,
					ChannelID: TestChannelID,
					Content:   TestInboundMessageBody,
				},
				Metadata: Metadata{
					Source:      "gateway",
					Dest:        "",
					ID:          uuid.FromStringOrNil(TestInboundMetadataUUID),
					Reply:       false, // Inbound message is not a reply
					InReplyTo:   "",    // Inbound message is not a reply
					MentionUser: false, // Inbound message does not mention the user
				},
			},
			want: &MessageSend{
				Content:   TestOutboundMessageBody,
				ChannelID: TestChannelID, // ChannelID should be the same as the original message
				Metadata: Metadata{
					Source:      TestAppName,                                    // Source should be the app sending the response
					Dest:        TestMetdataSource,                              // Dest should be the source from the original message
					ID:          uuid.FromStringOrNil(TestOutboundMetadataUUID), // ID should be a new UUID
					Reply:       true,                                           // Outbound message should be a reply to the original message
					InReplyTo:   TestInboundDiscordMessageID,                    // Outbound message should be a reply to the original message
					MentionUser: true,                                           // Outbound message should not mention the user
				},
			},
			sourceApp:     TestAppName,
			content:       TestOutboundMessageBody,
			wantErr:       false,
			shouldReply:   true,
			shouldMention: true,
		},
	}

	// Iterate through the test cases
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Create a new MessageSend struct
			got := tt.message.RespondToChannelOrThread(tt.sourceApp, tt.content, tt.shouldReply, tt.shouldMention)

			// Setup a filter to ignore the ID field
			filter := diff.Filter(
				func(path []string, parent reflect.Type, field reflect.StructField) bool {
					return field.Name != "ID"
				})

			changelog, err := diff.Diff(tt.want, got, filter)
			if err != nil {
				t.Errorf("Message_RespondToChannelOrThread() error = %v", err)
				return
			}

			// If the changelog is not empty, the test has failed
			if len(changelog) != 0 {
				t.Errorf("Message.RespondToChannelOrThread()")
				// Print the changelog to the console
				for _, change := range changelog {
					t.Errorf("Field %s changed from %v to %v", change.Path, change.From, change.To)
				}
			}
		})
	}
}
