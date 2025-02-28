package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminData struct {
	Username string `json:"Username"`
}

type FileData struct {
	Admins          map[string]AdminData `json:"Admins"`
	AuthorizedUsers map[string]bool      `json:"AuthorizedUsers"`
}

type MessageHandler struct {
	adminCode       string
	admins          map[string]bool
	authorizedUsers map[string]bool
	mu              sync.RWMutex
	adminFile       string
}

var Messages map[string]string

func LoadMessages() {

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	dataDir := os.Getenv("DATA_DIR")

	if dataDir == "" {
		dataDir = "data"
	}

	projectRoot := filepath.Dir(wd)
	messagesPath := filepath.Join(projectRoot, dataDir, "messages.json")

	absPath, _ := filepath.Abs(messagesPath)
	log.Printf("Attempting to read file at: %s", absPath)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Fatalf("File does not exist: %s", absPath)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		log.Fatalf("Error reading messages file: %v, filepath: %v", err, dataDir)
	}

	err = json.Unmarshal(data, &Messages)
	if err != nil {
		log.Fatalf("Error unmarshaling messages: %v", err)
	}
}

func getMessage(key string, args ...interface{}) string {
	msg, ok := Messages[key]
	if !ok {
		log.Printf("Warning: message key '%s' not found", key)
		return "Message not found"
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

func NewMessageHandler(adminCode string, adminFile string) *MessageHandler {
	h := &MessageHandler{
		adminCode:       adminCode,
		admins:          make(map[string]bool),
		authorizedUsers: make(map[string]bool),
		adminFile:       adminFile,
	}
	if err := h.loadData(); err != nil {
		log.Printf("Could not load admin data: %v", err)
		h.admins = make(map[string]bool)
		h.authorizedUsers = make(map[string]bool)
	}
	return h
}

func (h *MessageHandler) loadData() error {
	data, err := os.ReadFile(h.adminFile)
	if err != nil {
		if os.IsNotExist(err) {
			h.admins = make(map[string]bool)
			h.authorizedUsers = make(map[string]bool)
			fmt.Errorf("error reading admin file: %v", err)
			return nil
		}
		return err
	}

	var fileData FileData
	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("error unmarshaling data: %v", err)
	}
	if fileData.Admins == nil {
		fileData.Admins = make(map[string]AdminData)
	}
	if fileData.AuthorizedUsers == nil {
		fileData.AuthorizedUsers = make(map[string]bool)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.admins = make(map[string]bool)
	for _, adminData := range fileData.Admins {
		h.admins[adminData.Username] = true
	}

	h.authorizedUsers = fileData.AuthorizedUsers

	return nil
}

func (h *MessageHandler) saveData() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	fileData := FileData{
		Admins:          make(map[string]AdminData),
		AuthorizedUsers: h.authorizedUsers,
	}

	for username := range h.admins {
		fileData.Admins[username] = AdminData{
			Username: username,
		}
	}

	data, err := json.MarshalIndent(fileData, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling data: %v", err)
	}

	return os.WriteFile(h.adminFile, data, 0644)
}

func (h *MessageHandler) HandleCommand(msg *tgbotapi.Message) (*tgbotapi.MessageConfig, error) {
	cmd := msg.Command()
	args := msg.CommandArguments()

	switch cmd {
	case "register_admin":
		return h.handleRegisterAdmin(msg, args)
	case "authorize_user":
		return h.handleAuthorizeUser(msg, args)
	case "deauthorize_user":
		return h.handleDeauthorizeUser(msg, args)
	case "add_to_guide":
		return h.handleAddToGuide(msg)
	case "help":
		return h.handleHelp(msg)
	default:
		return h.createMessage(msg.Chat.ID, "Unknown command. Use /help to see available commands."), nil
	}
}

func (h *MessageHandler) handleRegisterAdmin(msg *tgbotapi.Message, code string) (*tgbotapi.MessageConfig, error) {
	if code == "" {
		return h.createMessage(msg.Chat.ID, getMessage("registerAdminUsage")), nil
	}

	if code != h.adminCode {
		return h.createMessage(msg.Chat.ID, getMessage("invalidAdminCode")), nil
	}

	h.mu.Lock()
	h.admins[msg.From.UserName] = true
	h.mu.Unlock()

	if err := h.saveData(); err != nil {
		log.Printf("Failed to save admin data: %v", err)
	}

	return h.createMessage(msg.Chat.ID, getMessage("adminRegistered")), nil
}

func (h *MessageHandler) handleAuthorizeUser(msg *tgbotapi.Message, username string) (*tgbotapi.MessageConfig, error) {
	h.mu.RLock()
	isAdmin := h.admins[msg.From.UserName]
	h.mu.RUnlock()

	if !isAdmin {
		return h.createMessage(msg.Chat.ID, getMessage("unauthorizedCommand")), nil
	}

	username = strings.TrimPrefix(username, "@")
	if username == "" {
		return h.createMessage(msg.Chat.ID, getMessage("authorizeUserUsage")), nil
	}

	h.mu.Lock()
	h.authorizedUsers[username] = true
	h.mu.Unlock()

	if err := h.saveData(); err != nil {
		log.Printf("Failed to save authorized users data: %v", err)
	}

	return h.createMessage(msg.Chat.ID, getMessage("userAuthorized", username)), nil
}

func (h *MessageHandler) handleDeauthorizeUser(msg *tgbotapi.Message, username string) (*tgbotapi.MessageConfig, error) {
	h.mu.RLock()
	isAdmin := h.admins[msg.From.UserName]
	h.mu.RUnlock()

	if !isAdmin {
		return h.createMessage(msg.Chat.ID, getMessage("unauthorizedCommand")), nil
	}

	username = strings.TrimPrefix(username, "@")
	if username == "" {
		return h.createMessage(msg.Chat.ID, getMessage("deauthorizeUserUsage")), nil
	}

	h.mu.Lock()
	delete(h.authorizedUsers, username)
	h.mu.Unlock()

	if err := h.saveData(); err != nil {
		log.Printf("Failed to save authorized users data: %v", err)
	}

	return h.createMessage(msg.Chat.ID, getMessage("userDeauthorized", username)), nil
}

func (h *MessageHandler) handleAddToGuide(msg *tgbotapi.Message) (*tgbotapi.MessageConfig, error) {
	h.mu.RLock()
	isAuthorized := h.authorizedUsers[msg.From.UserName]
	h.mu.RUnlock()

	if !isAuthorized {
		return h.createMessage(msg.Chat.ID, getMessage("unauthorizedCommand")), nil
	}

	if msg.ReplyToMessage == nil {
		return h.createMessage(msg.Chat.ID, getMessage("replyRequired")), nil
	}

	return h.createMessage(msg.Chat.ID, getMessage("thankYou", msg.From.UserName)), nil
}

func (h *MessageHandler) handleHelp(msg *tgbotapi.Message) (*tgbotapi.MessageConfig, error) {
	helpText := `Available commands:

/register_admin - Register as an admin using the admin code
/authorize_user - Grant a user permission to submit guide entries (admin only)
/deauthorize_user - Remove a user's permission to submit entries (admin only)
/add_to_guide - Submit a message to be added to the guide (reply to a message)

For authorized users: Reply to any message with /add_to_guide to submit it for the guide.
For admins: Use @username format when authorizing or deauthorizing users.`

	return h.createMessage(msg.Chat.ID, helpText), nil
}

func (h *MessageHandler) createMessage(chatID int64, text string) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)
	return &msg
}
