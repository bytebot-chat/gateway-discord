package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/r3labs/diff/v3"
	uuid "github.com/satori/go.uuid"
)

func TestMessage_RespondToChannelOrThread(t *testing.T) {
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
					Source: TestMetdataSource,                             // Inbound messages should always have a source or else no app will know where to send responses
					Dest:   TestMetdataDest,                               // Inbound messages typically will not have a destination
					ID:     uuid.FromStringOrNil(TestInboundMetadataUUID), // Usually this is set by the app, but we can set it here for testing
				},
			},
			want: &MessageSend{
				Content:          TestOutboundMessageBody,
				ChannelID:        TestChannelID,
				MessageReference: discordgo.MessageReference{}, // Should be empty because we're not replying to anything
				Metadata: Metadata{
					Source: TestAppName,       // Should be the app name
					Dest:   TestMetdataSource, // Outbound messages should always have a destination or else no app will know to process them
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
					ID:        TestInboundDiscordMessageID, // This is the ID of the message we will be replying to.
					ChannelID: TestChannelID,               // This is the ID of the channel the message originated from (or the thread if it was a thread).
					Content:   TestInboundMessageBody,
					GuildID:   TestGuildID,
				},
				Metadata: Metadata{
					Source: TestMetdataSource,                             // Inbound messages should always have a source or else no app will know where to send responses
					Dest:   TestMetdataDest,                               // Inbound messages typically will not have a destination but we can set it here for testing
					ID:     uuid.FromStringOrNil(TestInboundMetadataUUID), // Usually this is set by the app, but we can set it here for testing
				},
			},
			want: &MessageSend{
				Content:   TestOutboundMessageBody, // This is the text we want to send in the response
				ChannelID: TestChannelID,           // ChannelID should be the same as the original message. Comes from Message.ChannelID.
				MessageReference: discordgo.MessageReference{
					MessageID: TestInboundDiscordMessageID, // This is the ID of the message we are replying to.
					ChannelID: TestChannelID,               // This is the ID of the channel the message originated from (or the thread if it was a thread).
					GuildID:   TestGuildID,                 // This is the ID of the guild the message originated from.
				},
				Metadata: Metadata{
					Source: TestAppName,                                    // Source should be the app sending the response
					Dest:   TestMetdataSource,                              // Dest should be the source from the original message. Comes from Message.Metadata.Source.
					ID:     uuid.FromStringOrNil(TestOutboundMetadataUUID), // ID should be a new UUID v4
				},
			},
			sourceApp:     TestAppName,
			content:       TestOutboundMessageBody,
			wantErr:       false,
			shouldReply:   true,  // This should be true because we are replying to a message
			shouldMention: false, // This should be false because we are not mentioning the original message author
		},
		{
			name: "Reply with mention",
			message: &Message{
				Message: &discordgo.Message{
					ID:        TestInboundDiscordMessageID,
					ChannelID: TestChannelID,
					Content:   TestInboundMessageBody,
					GuildID:   TestGuildID,
				},
				Metadata: Metadata{
					Source: TestMetdataSource,
					Dest:   TestMetdataDest,
					ID:     uuid.FromStringOrNil(TestInboundMetadataUUID),
				},
			},
			want: &MessageSend{
				Content:   TestOutboundMessageBody,
				ChannelID: TestChannelID, // ChannelID should be the same as the original message
				MessageReference: discordgo.MessageReference{
					MessageID: TestInboundDiscordMessageID,
					ChannelID: TestChannelID,
					GuildID:   TestGuildID,
				},
				Metadata: Metadata{
					Source: TestAppName,                                    // Source should be the app sending the response
					Dest:   TestMetdataSource,                              // Dest should be the source from the original message
					ID:     uuid.FromStringOrNil(TestOutboundMetadataUUID), // ID should be a new UUID
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
				// Print the changelog to the console
				for _, c := range changelog {
					t.Errorf("Message.RespondToChannelOrThread() - %s\nCompare this snippet from %s:\nWanted:\t%v\nGot:\t%v\n", tt.name, strings.Join(c.Path, "."), c.From, c.To)
				}
			}
		})
	}
}

func TestMessageSend_UnmarshalJSON(t *testing.T) {
	type fields struct {
		ChannelID string
		Content   string
		Metadata  Metadata
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Valid JSON",
			fields: fields{
				ChannelID: TestChannelID,
				Content:   TestOutboundMessageBody,
				Metadata: Metadata{
					Source: TestAppName,
					Dest:   TestMetdataSource, // Outbound messages should always have a destination or else no app will know to process them
					ID:     uuid.FromStringOrNil(TestOutboundMetadataUUID),
				},
			},
			args: args{
				b: []byte(`{
					"channel_id": "` + TestChannelID + `",
					"content": "` + TestOutboundMessageBody + `",
					"metadata": {
						"source": "` + TestAppName + `",
						"dest": "` + TestMetdataSource + `",
						"id": "` + TestOutboundMetadataUUID + `"
					}
				}`),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MessageSend{
				ChannelID: tt.fields.ChannelID,
				Content:   tt.fields.Content,
				Metadata:  tt.fields.Metadata,
			}
			if err := m.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("MessageSend.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageSend_MarshalJSON(t *testing.T) {
	type fields struct {
		ChannelID string
		Content   string
		Metadata  Metadata
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "Valid JSON",
			fields: fields{
				ChannelID: TestChannelID,
				Content:   TestOutboundMessageBody,
				Metadata: Metadata{
					Source: TestAppName,
					Dest:   TestMetdataSource, // Outbound messages should always have a destination or else no app will know to process them
					ID:     uuid.FromStringOrNil(TestOutboundMetadataUUID),
				},
			},
			want:    []byte(`{"channel_id":"` + TestChannelID + `","content":"` + TestOutboundMessageBody + `","metadata":{"source":"` + TestAppName + `","dest":"` + TestMetdataSource + `","id":"` + TestOutboundMetadataUUID + `"}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MessageSend{
				ChannelID: tt.fields.ChannelID,
				Content:   tt.fields.Content,
				Metadata:  tt.fields.Metadata,
			}
			got, err := m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MessageSend.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageSend.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
