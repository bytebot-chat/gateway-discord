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
	TestChannelID                = "000000000000000000"
	TestInboundMetadataUUID      = "00000000-0000-0000-0000-000000000000"
	TestOutboundMetadataUUID     = "11111111-1111-1111-1111-111111111111"
	TestInboundDiscordMessageID  = "222222222222222222"
	TestOutboundDiscordMessageID = "333333333333333333"
	TestInboundMessageBody       = "hello world"
	TestOutboundMessageBody      = "goodbye world"
	TestUserID                   = "000000000000000000"
	TestUserName                 = "test-user"
	TestUserDiscriminator        = "0000"
	TestMetdataSource            = "gateway"
	TestMetdataDest              = ""
	TestAppName                  = "test-app"
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
		// Super basic test case
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
					Content:   TestInboundMessageBody,
					ChannelID: TestChannelID,
					Author: &discordgo.User{
						ID:            TestUserID,
						Username:      TestUserName,
						Discriminator: TestUserDiscriminator,
					},
				},
				Metadata: Metadata{
					Source: TestMetdataSource,
					Dest:   "",
					ID:     uuid.FromStringOrNil(TestOutboundMetadataUUID),
				},
			},
			wantErr: false,
		},
		// Full message test case
		{
			name: "Full message body",
			messageJSON: []byte(`
			{
				"metadata": {},
				"message": {
					"id": "1041032867578396712",
					"channel_id": "1037905587968688178",
					"guild_id": "1037896477197996182",
					"content": "test",
					"timestamp": "2022-11-12T16:52:57.086Z",
					"edited_timestamp": null,
					"mention_roles": [],
					"tts": false,
					"mention_everyone": false,
					"author": {
						"id": "179258058118135808",
						"email": "",
						"username": "fraq",
						"avatar": "1ac33c0aa68e5fd0ce5a22f7d1e02b22",
						"locale": "",
						"discriminator": "1337",
						"token": "",
						"verified": false,
						"mfa_enabled": false,
						"banner": "",
						"accent_color": 0,
						"bot": false,
						"public_flags": 0,
						"premium_type": 0,
						"system": false,
						"flags": 0
					},
					"attachments": [],
					"embeds": [],
					"mentions": [],
					"reactions": null,
					"pinned": false,
					"type": 0,
					"webhook_id": "",
					"member": {
						"guild_id": "",
						"joined_at": "2022-11-04T01:10:03.358Z",
						"nick": "",
						"deaf": false,
						"mute": false,
						"avatar": "",
						"user": null,
						"roles": [],
						"premium_since": null,
						"pending": false,
						"permissions": "0",
						"communication_disabled_until": null
					},
					"mention_channels": null,
					"activity": null,
					"application": null,
					"message_reference": null,
					"referenced_message": null,
					"interaction": null,
					"flags": 0,
					"sticker_items": null
				}
			}

			`),
			testCase: &Message{
				Message:  &discordgo.Message{},
				Metadata: Metadata{},
			},
			want: &Message{
				Message: &discordgo.Message{
					ID:              TestInboundDiscordMessageID,
					ChannelID:       TestChannelID,
					GuildID:         TestChannelID,
					Content:         TestInboundMessageBody,
					MentionRoles:    []string{},
					MentionEveryone: false,
					Author: &discordgo.User{
						ID:            TestUserID,
						Email:         "",
						Avatar:        "",
						Locale:        "",
						Discriminator: "1337",
						Token:         "",
						Verified:      false,
						Banner:        "",
						AccentColor:   0,
						Bot:           false,
						PublicFlags:   0,
						PremiumType:   0,
						System:        false,
						Flags:         0,
					},
					Reactions: nil,
					Pinned:    false,
					Type:      0,
					WebhookID: "",
					Member: &discordgo.Member{
						GuildID: "",
						Nick:    "",
						Deaf:    false,
						Mute:    false,
						Avatar:  "",
						User:    nil,
						Roles:   []string{},
						Pending: false,
					},
					MentionChannels:   nil,
					Activity:          nil,
					Application:       nil,
					MessageReference:  nil,
					ReferencedMessage: nil,
					Interaction:       nil,
					Flags:             0,
					StickerItems:      nil,
				},
				Metadata: Metadata{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testCase.UnmarshalJSON(tt.messageJSON)
			if (err != nil) != tt.wantErr {
				fmt.Println(string(tt.messageJSON)) // Print the JSON for debugging

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
