package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	TelegramToken string
	AdminCode     string
	AdminFile     string
	MessagesFile  string
}

// Bot represents the Telegram bot instance
type Bot struct {
	api            *tgbotapi.BotAPI
	messageHandler *MessageHandler
}

// New creates a new Bot instance
func New(config *Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	log.Printf("Config: %v", config)

	return &Bot{
		api:            api,
		messageHandler: NewMessageHandler(config.AdminCode, config.AdminFile),
	}, nil
}

// Start begins polling for updates
func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	log.Printf("Bot started. Authorized as %s", b.api.Self.UserName)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			reply, err := b.messageHandler.HandleCommand(update.Message)
			if err != nil {
				log.Printf("Error handling command: %v", err)
				continue
			}
			if reply != nil {
				if _, err := b.api.Send(reply); err != nil {
					log.Printf("Error sending message: %v", err)
				}
			}
		}
	}

	return nil
}
