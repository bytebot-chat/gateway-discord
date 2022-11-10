package model

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/r3labs/diff"
	uuid "github.com/satori/go.uuid"
)

func TestMessage_Unmarshal(t *testing.T) {

	tests := []struct {
		name         string
		discordJSON  []byte
		metadataJSON []byte
		want         *Message
		testCase     *Message
		wantErr      bool
	}{
		{
			name: "hello world",
			discordJSON: []byte(`
			{
				"content": "hello world"
			}
			`),
			metadataJSON: []byte{},
			testCase: &Message{
				Message:  &discordgo.Message{},
				Metadata: Metadata{},
			},
			want: &Message{
				Message: &discordgo.Message{
					Content: "hello world",
				},
				Metadata: Metadata{},
			},
			wantErr: false,
		},
		{
			name: "hello world with metadata",
			discordJSON: []byte(`
			{
				"content": "hello world"
			}
			`),
			metadataJSON: []byte(`
			{
				"source": "source-app",
				"dest": "dest-app",
				"id": "00000000-0000-0000-0000-000000000000"
			}
			`),
			want: &Message{
				Message: &discordgo.Message{
					Content: "hello world",
				},
				Metadata: Metadata{
					Source: "source-app",
					Dest:   "dest-app",
					ID:     uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000"),
				},
			},
			testCase: &Message{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Unmarshal the discord json into the message and check for deserialization errors
			if err := tt.testCase.Unmarshal(tt.discordJSON); (err != nil) != tt.wantErr {
				t.Errorf("Message.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Compare the testCase to the want value
			if diff.Changed(tt.testCase.Message, tt.want.Message) {
				d, _ := diff.Diff(tt.testCase, tt.want)
				fmt.Printf("\n--------\nI have:\n\n%+v\n", tt.testCase.Message)
				fmt.Printf("\n--------\nI want:\n\n%+v\n", tt.want.Message)
				t.Errorf("Message.Unmarshal() = test case does not match want value: %s", d)
			}
		})
	}

}
