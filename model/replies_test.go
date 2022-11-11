package model

import (
	"encoding/json"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/r3labs/diff"
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
		name      string       // name of the test
		message   *Message     // Original message to respond to
		want      *MessageSend // Expected response
		sourceApp string       // ID of the app sending the response
		content   string       // Text to send in the response
		wantErr   bool         // Whether or not the test should fail
	}{
		{
			name: "hello world",
			message: &Message{
				Message: &discordgo.Message{
					Content:   "hello world",
					ChannelID: "000000000000000000",
				},
				Metadata: Metadata{
					Source: "gateway",
					Dest:   "",
				},
			},
			want: &MessageSend{
				Content:   "hello world",
				ChannelID: "000000000000000000",
				Metadata: Metadata{
					Source: "test-app",
					Dest:   "gateway",
				},
			},
			sourceApp: "test-app",
			content:   "hello world",
			wantErr:   false,
		},
	}

	// Iterate through the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				got      MessageSend
				gotBytes []byte
				err      error
			)

			// Execute the function we're testing
			gotBytes, err = tt.message.RespondToChannelOrThread(tt.sourceApp, tt.content)

			// Check for errors
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.RespondToChannelOrThread() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Unmarshal the bytes into the MessageSend struct for comparison
			// We have to do this because the function we're testing returns a byte slice
			// and we want to compare the struct values
			err = json.Unmarshal(gotBytes, &got)
			if err != nil {
				t.Errorf("Message.RespondToChannelOrThread() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Set the ID of the MessageSend struct to the ID we generated
			got.Metadata.ID = uuid.FromStringOrNil(TestMetadataUUID)
			tt.want.Metadata.ID = uuid.FromStringOrNil(TestMetadataUUID)

			// Compare what we got to what we want from the test case
			if diff.Changed(got, tt.want) {
				t.Errorf("Message.RespondToChannelOrThread()\n--------\ngot:\n\n%+v\n\n--------\nwanted:\n\n%+v\n", string(gotBytes), tt.want)
			}

		})
	}
}
