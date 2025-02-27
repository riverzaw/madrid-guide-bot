package bot

import (
	"encoding/json"
	"os"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func setupTestFile(t *testing.T, initialData string) (string, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "admins_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if err := os.WriteFile(tmpFile.Name(), []byte(initialData), 0644); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	return tmpFile.Name(), func() {
		os.Remove(tmpFile.Name())
	}
}

const (
	testChatID    = 12345
	testAdminData = `{"Admins": {"adminUser": {"Username": "adminUser", "ChatID": 0}}, "AuthorizedUsers": {"authUser": true}}`
	emptyData     = `{}`
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
			tmpFile, cleanup := setupTestFile(t, emptyData)
			defer cleanup()

			h := NewMessageHandler(tt.adminCode, tmpFile)

			response, err := h.handleRegisterAdmin(tt.message, tt.message.CommandArguments())

			if (err != nil) != tt.expectedError {
				t.Errorf("handelRegisterAdmin() error = %v, expectedError %v", err, tt.expectedError)
			}

			if response.Text != tt.expectedReply {
				t.Errorf("Expected reply %q, got %q", tt.expectedReply, response.Text)
			}

			h.mu.RLock()
			_, isAdmin := h.admins[tt.message.From.UserName]
			h.mu.RUnlock()

			if isAdmin != tt.expectedAdmin {
				t.Errorf("Expected admin status %v, got %v", tt.expectedAdmin, isAdmin)
			}

			if tt.expectedAdmin {
				data, err := os.ReadFile(tmpFile)
				if err != nil {
					t.Fatalf("Failed to read admin file: %v", err)
				}

				var fileData FileData
				if err := json.Unmarshal(data, &fileData); err != nil {
					t.Fatalf("Failed to unmarshal admin data: %v", err)
				}

				if _, exists := fileData.Admins[tt.message.From.UserName]; !exists {
					t.Error("Admin was not properly saved to file")
				}
			}

		})
	}
}

func TestHandleAuthorizeUser(t *testing.T) {
	LoadMessages()

	tests := []struct {
		name               string
		adminCode          string
		message            *tgbotapi.Message
		username           string
		expectedReply      string
		expectedAuthorized bool
		expectedError      bool
	}{
		{
			name:      "Unauthorized User",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text: "/authorize_user @user1",
				From: &tgbotapi.User{UserName: "noAdmin"},
				Chat: &tgbotapi.Chat{ID: 12345},
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: 15,
					},
				},
			},
			username:           "user1",
			expectedReply:      getMessage("unauthorizedCommand"),
			expectedAuthorized: false,
		},
		{
			name:      "No argument usage",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text: "/authorize_user",
				From: &tgbotapi.User{UserName: "adminUser"},
				Chat: &tgbotapi.Chat{ID: 12345},
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: 15,
					},
				},
			},
			username:           "",
			expectedReply:      getMessage("authorizeUserUsage"),
			expectedAuthorized: false,
		},
		{
			name:      "Valid usage",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text: "/authorize_user @user1",
				From: &tgbotapi.User{UserName: "adminUser"},
				Chat: &tgbotapi.Chat{ID: 12345},
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: 15,
					},
				},
			},
			username:           "user1",
			expectedReply:      getMessage("userAuthorized", "user1"),
			expectedAuthorized: true,
		},
	}

	tmpFile, cleanup := setupTestFile(t, testAdminData)
	defer cleanup()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMessageHandler(tt.adminCode, tmpFile)

			response, err := h.handleAuthorizeUser(tt.message, tt.username)

			if (err != nil) != tt.expectedError {
				t.Errorf("handleAuthorizeUser() error = %v, expectedError %v", err, tt.expectedError)
			}

			if response.Text != tt.expectedReply {
				t.Errorf("Expected reply %q, got %q", tt.expectedReply, response.Text)
			}

			h.mu.RLock()
			_, isAuthorized := h.authorizedUsers[tt.username]
			h.mu.RUnlock()

			if isAuthorized != tt.expectedAuthorized {
				t.Errorf("Expected user status %v, got %v", tt.expectedAuthorized, isAuthorized)
			}

		})
	}
}

func TestHandleDeauthorizeUser(t *testing.T) {
	LoadMessages()

	tests := []struct {
		name               string
		adminCode          string
		isFromAdmin        bool
		message            *tgbotapi.Message
		username           string
		expectedReply      string
		expectedAuthorized bool
		expectedError      bool
	}{
		{
			name:      "Unauthorized user",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text: "/deauthorize_user @authUser",
				From: &tgbotapi.User{UserName: "noAdmin"},
				Chat: &tgbotapi.Chat{ID: 12345},
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: 17,
					},
				},
			},
			username:           "authUser",
			expectedReply:      getMessage("unauthorizedCommand"),
			expectedAuthorized: true,
		},
		{
			name:      "No argument usage",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text: "/deauthorize_user",
				From: &tgbotapi.User{UserName: "adminUser"},
				Chat: &tgbotapi.Chat{ID: 12345},
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: 17,
					},
				},
			},
			username:           "",
			expectedReply:      getMessage("deauthorizeUserUsage"),
			expectedAuthorized: false,
		},
		{
			name:      "Valid usage",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text: "/deauthorize_user @authUser",
				From: &tgbotapi.User{UserName: "adminUser"},
				Chat: &tgbotapi.Chat{ID: 12345},
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: 17,
					},
				},
			},
			username:           "authUser",
			expectedReply:      getMessage("userDeauthorized", "authUser"),
			expectedAuthorized: false,
		},
	}

	tmpFile, cleanup := setupTestFile(t, testAdminData)
	defer cleanup()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMessageHandler(tt.adminCode, tmpFile)

			response, err := h.handleDeauthorizeUser(tt.message, tt.username)

			if (err != nil) != tt.expectedError {
				t.Errorf("handleDeauthorizeUser() error = %v, expectedError %v", err, tt.expectedError)
			}

			if response.Text != tt.expectedReply {
				t.Errorf("Expected reply %q, got %q", tt.expectedReply, response.Text)
			}

			h.mu.RLock()
			_, isAuthorized := h.authorizedUsers[tt.username]
			h.mu.RUnlock()

			if isAuthorized != tt.expectedAuthorized {
				t.Errorf("Expected user status %v, got %v", tt.expectedAuthorized, isAuthorized)
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
			name:      "Unathorized usage",
			adminCode: "admin123",
			message: &tgbotapi.Message{
				Text: "/add_to_guide",
				From: &tgbotapi.User{UserName: "unauthUser"},
				Chat: &tgbotapi.Chat{ID: 12345},
				Entities: []tgbotapi.MessageEntity{{
					Type:   "bot_command",
					Offset: 0,
					Length: 12,
				}},
			},
			username:      "unauthUser",
			expectedReply: getMessage("unauthorizedCommand"),
		},
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

	tmpFile, cleanup := setupTestFile(t, testAdminData)
	defer cleanup()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			h := NewMessageHandler(tt.adminCode, tmpFile)

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
