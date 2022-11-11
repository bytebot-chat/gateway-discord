package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/r3labs/diff"
	uuid "github.com/satori/go.uuid"
)

// Values used in tests
const (
	TestChannelID         = "000000000000000000"
	TestMetadataUUID      = "00000000-0000-0000-0000-000000000000"
	TestUserID            = "000000000000000000"
	TestUserName          = "test-user"
	TestUserDiscriminator = "0000"
	TestMetdataSource     = "gateway"
	TestMetdataDest       = ""
)

func TestMessage_UnmarshalJSON(t *testing.T) {
	/*
		This test case checks that the UnmarshalJSON method returns the correct Message struct
		The function is intended to unmarshal a JSON string into a Message struct
		This means the Message struct should have the same values as the JSON string

		Because the *discordgo.Message struct is embedded in the Message struct and also has a MarshalJSON method,
		go will call the MarshalJSON method of the *discordgo.Message struct when the Message struct is marshaled
		unless we override it with a custom MarshalJSON method in the Message struct, which we do

	*/

	tests := []struct {
		name        string
		messageJSON []byte
		want        *Message
		testCase    *Message
		wantErr     bool
	}{
		{
			name: "hello world",
			messageJSON: []byte(`{
				"metadata": {
					"source": "gateway",
					"dest": "",
					"id": "00000000-0000-0000-0000-000000000000"
				},
				"message": {
					"content": "hello world",
					"channel_id": "000000000000000000",
					"author": {
						"id": "000000000000000000",
						"username": "test-user",
						"discriminator": "0000"
					}
				}
			}`),
			testCase: &Message{
				Message:  &discordgo.Message{},
				Metadata: Metadata{},
			},
			want: &Message{
				Message: &discordgo.Message{
					Content:   "hello world",
					ChannelID: TestChannelID,
					Author: &discordgo.User{
						ID:            TestUserID,
						Username:      TestUserName,
						Discriminator: TestUserDiscriminator,
					},
				},
				Metadata: Metadata{
					Source: "gateway",
					Dest:   "",
					ID:     uuid.FromStringOrNil(TestMetadataUUID),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testCase.UnmarshalJSON(tt.messageJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare content of messages
			if diff.Changed(tt.testCase.Message, tt.want.Message) {
				t.Errorf("Message.UnmarshalJSON() Message does not match")

				d, err := diff.Diff(tt.testCase.Message, tt.want.Message)
				if err != nil {
					t.Errorf("Message.UnmarshalJSON() Diff error = %v", err)
				}

				for _, c := range d {
					fmt.Printf("Compare this snippet from %s:\n", strings.Join(c.Path, "."))
					fmt.Printf("Want: %v\n", c.From)
					fmt.Printf("Got:  %v\n", c.To)
				}
			}

			// Compare metadata
			if diff.Changed(tt.testCase.Metadata, tt.want.Metadata) {
				t.Errorf("Message.UnmarshalJSON() Metadata does not match")

				d, err := diff.Diff(tt.testCase.Metadata, tt.want.Metadata)
				if err != nil {
					t.Errorf("Message.UnmarshalJSON() Diff error = %v", err)
				}

				for _, c := range d {
					fmt.Printf("Compare this snippet from %s:\n", strings.Join(c.Path, "."))
					fmt.Printf("Want: %v\n", c.From)
					fmt.Printf("Got:  %v\n", c.To)
				}
			}
		})
	}
}
