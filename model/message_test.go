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
	TestChannelID                = "test-channel-id"
	TestInboundMetadataUUID      = "00000000-0000-0000-0000-000000000000"
	TestOutboundMetadataUUID     = "11111111-1111-1111-1111-111111111111"
	TestInboundDiscordMessageID  = "test-inbound-discord-message-id"
	TestOutboundDiscordMessageID = "test-outbound-discord-message-id"
	TestInboundMessageBody       = "test-inbound-message-body"
	TestOutboundMessageBody      = "test-outbound-message-body"
	TestUserID                   = "test-user-id"
	TestUserName                 = "test-user-name"
	TestUserDiscriminator        = "0000"
	TestMetdataSource            = "test-source"
	TestMetdataDest              = "test-dest"
	TestAppName                  = "test-app"
	TestGuildID                  = "test-guild-id"
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
					"source": "` + TestMetdataSource + `",
					"dest": "",
					"id": "00000000-0000-0000-0000-000000000000"
				},
				"message": {
					"content": "` + TestInboundMessageBody + `",
					"channel_id": "` + TestChannelID + `",
					"author": {
						"id": "` + TestUserID + `",
						"username": "` + TestUserName + `",
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
					"id": "` + TestInboundDiscordMessageID + `",
					"channel_id": "` + TestChannelID + `",
					"guild_id": "` + TestGuildID + `",
					"content": "` + TestInboundMessageBody + `",
					"edited_timestamp": null,
					"mention_roles": [],
					"tts": false,
					"mention_everyone": false,
					"author": {
						"id": "` + TestUserID + `",
						"email": "",
						"username": "` + TestUserName + `",
						"avatar": "",
						"locale": "",
						"discriminator": "` + TestUserDiscriminator + `",
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
					GuildID:         TestGuildID,
					Content:         TestInboundMessageBody,
					MentionRoles:    []string{},
					MentionEveryone: false,
					Author: &discordgo.User{
						ID:            TestUserID,
						Email:         "",
						Avatar:        "",
						Locale:        "",
						Discriminator: TestUserDiscriminator,
						Token:         "",
						Verified:      false,
						Banner:        "",
						AccentColor:   0,
						Bot:           false,
						PublicFlags:   0,
						PremiumType:   0,
						System:        false,
						Flags:         0,
						Username:      TestUserName,
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
					t.Errorf("Compare this snippet from %s:\nWanted:\t%v\nGot:\t%v\n", strings.Join(c.Path, "."), c.From, c.To)
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
					t.Errorf("Compare this snippet from %s:\nWanted:\t%v\nGot:\t%v\n", strings.Join(c.Path, "."), c.From, c.To)
				}
			}
		})
	}
}
