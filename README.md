# Telegram Scheduled Posting Bot

This is a Telegram bot that schedules and sends messages at a later time.

## Features

- Schedule a message to be posted later.
- Uses PostgreSQL for storing scheduled posts.
- Built with Go and uses the go-telegram/bot library.
- Dockerized for easy deployment.

## Requirements

- Go 1.24.5
- PostgreSQL
- Telegram Bot token from @BotFather

## Getting Started

- Clone the repository
- Set environment variables in .env file:

BOT_TOKEN: Telegram bot token<br>
DB_USER: PostgreSQL username<br>
DB_PASS: PostgreSQL password<br>
DB_NAME: PostgreSQL database name<br>
DB_HOST: PostgreSQL host<br>
DB_PORT: PostgreSQL port<br>

- Run the bot using Docker Compose:

```bash
docker-compose up --build
```

- Interact with the bot in Telegram to schedule posts.

## How It Works

- When a user sends a message to the bot, it is scheduled to be posted later.
- The bot stores the message metadata (chat ID, from ID, message ID, and send time) in a PostgreSQL database.
- A background worker periodically checks the database for messages scheduled to be sent and sends them using CopyMessage.
