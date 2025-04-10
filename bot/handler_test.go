package bot

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestHandlerRegisterAdmin(t *testing.T) {
	LoadMessages()

	tests := []struct {
		name          string
		adminCode     string
		message       *tgbotapi.Message
		expectedAdmin bool
		expectedReply string
		expectedError bool
	}{
		{
			name:      "Valid registration",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text: "/register_admin admin123",
				From: &tgbotapi.User{UserName: "testUser"},
				Chat: &tgbotapi.Chat{ID: 12345},
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: 15,
					},
				},
			},
			expectedAdmin: true,
			expectedReply: getMessage("adminRegistered"),
		},
		{
			name:      "Invalid code",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text: "/register_admin wrongcode",
				From: &tgbotapi.User{UserName: "wrongUser"},
				Chat: &tgbotapi.Chat{ID: 12345},
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: 15,
					},
				},
			},
			expectedAdmin: false,
			expectedReply: getMessage("invalidAdminCode"),
		},
		{
			name:      "Missing code",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text: "/register_admin",
				From: &tgbotapi.User{UserName: "missingCodeUser"},
				Chat: &tgbotapi.Chat{ID: 12345},
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: 15,
					},
				},
			},
			expectedAdmin: false,
			expectedReply: getMessage("registerAdminUsage"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &tgbotapi.BotAPI{}

			h := NewMessageHandler(tt.adminCode, api)

			response, err := h.handleRegisterAdmin(tt.message, tt.message.CommandArguments())

			if (err != nil) != tt.expectedError {
				t.Errorf("handleRegisterAdmin() error = %v, expectedError %v", err, tt.expectedError)
			}

			if response.Text != tt.expectedReply {
				t.Errorf("Expected reply %q, got %q", tt.expectedReply, response.Text)
			}
		})
	}
}

func TestHandlerAddToGuide(t *testing.T) {
	LoadMessages()

	tests := []struct {
		name          string
		adminCode     string
		message       *tgbotapi.Message
		username      string
		expectedReply string
		expectedError bool
	}{
		{
			name:      "Reply required",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text: "/add_to_guide",
				From: &tgbotapi.User{UserName: "authUser"},
				Chat: &tgbotapi.Chat{ID: 12345},
				Entities: []tgbotapi.MessageEntity{{
					Type:   "bot_command",
					Offset: 0,
					Length: 12,
				}},
			},
			username:      "authUser",
			expectedReply: getMessage("replyRequired"),
		},
		{
			name:      "Successful usage",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text:           "/add_to_guide",
				From:           &tgbotapi.User{UserName: "authUser"},
				Chat:           &tgbotapi.Chat{ID: 12345},
				ReplyToMessage: &tgbotapi.Message{Text: "msg to add", From: &tgbotapi.User{UserName: "adminUser"}, Chat: &tgbotapi.Chat{ID: 12345}},
				Entities: []tgbotapi.MessageEntity{{
					Type:   "bot_command",
					Offset: 0,
					Length: 12,
				}},
			},
			username:      "authUser",
			expectedReply: getMessage("thankYou", "authUser"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &tgbotapi.BotAPI{}

			h := NewMessageHandler(tt.adminCode, api)

			response, err := h.handleAddToGuide(tt.message)

			if (err != nil) != tt.expectedError {
				t.Errorf("handleAddToGuide() error = %v, expectedError %v", err, tt.expectedError)
			}

			if response.Text != tt.expectedReply {
				t.Errorf("Expected reply %q, got %q", tt.expectedReply, response.Text)
			}
		})
	}
}
