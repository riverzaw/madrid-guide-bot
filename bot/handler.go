package bot

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strings"

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
	api             *tgbotapi.BotAPI
}

//go:embed messages.json
var MessagesJson []byte
var Messages map[string]string
var adminIDs = LoadAdmins()

func LoadMessages() {
	err := json.Unmarshal(MessagesJson, &Messages)
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

func NewMessageHandler(adminCode string, api *tgbotapi.BotAPI) *MessageHandler {
	h := &MessageHandler{
		adminCode:       adminCode,
		admins:          make(map[string]bool),
		authorizedUsers: make(map[string]bool),
		api:             api,
	}
	return h
}

func (h *MessageHandler) HandleCommand(msg *tgbotapi.Message) (*tgbotapi.MessageConfig, error) {
	cmd := msg.Command()
	args := msg.CommandArguments()

	switch cmd {
	case "register_admin":
		return h.handleRegisterAdmin(msg, args)
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

	addAdmin(msg, h.api)

	return h.createMessage(msg.Chat.ID, getMessage("adminRegistered")), nil
}

func (h *MessageHandler) handleAddToGuide(msg *tgbotapi.Message) (*tgbotapi.MessageConfig, error) {
	if msg.ReplyToMessage == nil {
		return h.createMessage(msg.Chat.ID, getMessage("replyRequired")), nil
	}

	for _, adminID := range adminIDs {
		contextMsg := getMessage("suggestionContext",
			msg.From.UserName, msg.Chat.Title, msg.Chat.ID)
		h.api.Send(tgbotapi.NewMessage(adminID, contextMsg))
		forwardedMsg := tgbotapi.NewForward(adminID, msg.ReplyToMessage.Chat.ID, msg.ReplyToMessage.MessageID)
		if _, err := h.api.Send(forwardedMsg); err != nil {
			log.Printf("Error forwarding message to admin %d: %v", adminID, err)
			continue
		}
	}

	return h.createMessage(msg.Chat.ID, getMessage("thankYou", msg.From.UserName)), nil
}

func (h *MessageHandler) handleHelp(msg *tgbotapi.Message) (*tgbotapi.MessageConfig, error) {
	helpText := `Available commands:

/register_admin - Register as an admin using the admin code
/add_to_guide - Submit a message to be added to the guide (reply to a message)
/help - Show help message
`

	return h.createMessage(msg.Chat.ID, helpText), nil
}

func (h *MessageHandler) createMessage(chatID int64, text string) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)
	return &msg
}

func addAdmin(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	userID := msg.From.ID
	if slices.Contains(adminIDs, userID) {
		bot.Send(tgbotapi.NewMessage(userID, "Ты уже админ!"))
		return
	}

	adminIDs = append(adminIDs, userID)
	saveAdminsToFlySecrets()

	bot.Send(tgbotapi.NewMessage(userID, "Теперь ты админ!"))
}

func saveAdminsToFlySecrets() {
	idsStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(adminIDs)), ","), "[]")

	cmd := exec.Command("fly", "secrets", "set", "ADMIN_IDS="+idsStr)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error updating Fly Secrets: %s", err)
	}
}
