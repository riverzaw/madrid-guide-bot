# Telegram Guide Bot

A Telegram bot that helps manage and collect guide entries through admin approval system.

## Features

- Admin registration
- Guide entry submission
- Multi-admin support

## Setup

1. Clone the repository
2. Copy `env.template` to `.env` and fill in your credentials:

   ```bash
   cp env.template .env
   ```

3. Build and run:

   ```bash
   go build -o madrid-guide-bot
   go run madrid-buide-bot
   ```

## Commands

- `/start` - Start the bot
- `/register_admin <code>` - Register as an admin
- `/add_to_guide` - Submit an entry to the guide

## Configuration

The bot uses the following environment variables:

- `TELEGRAM_BOT_TOKEN` - Your Telegram bot token
- `ADMIN_REGISTRATION_CODE` - Code for registering new admins

## License

MIT License - see LICENSE file
