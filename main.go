package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/riverzaw/madrid-guide-bot/bot"
)

func main() {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}

	config := &bot.Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		AdminCode:     os.Getenv("ADMIN_REGISTRATION_CODE"),
		AdminFile:     filepath.Join(dataDir, "admins.json"),
		MessagesFile:  filepath.Join(dataDir, "messages.json"),
	}

	if config.TelegramToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}
	if config.AdminCode == "" {
		log.Fatal("ADMIN_REGISTRATION_CODE environment variable is not set")
	}

	bot.LoadMessages()

	bot, err := bot.New(config)
	if err != nil {
		log.Fatal(err)
	}

	if err := bot.Start(); err != nil {
		log.Fatal(err)
	}
}
