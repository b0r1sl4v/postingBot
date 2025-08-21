package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type ScheduledPost struct {
	ChatID     string
	FromChatID string
	Message_id int
	SendAt     time.Time
}

var scheduledPosts []ScheduledPost
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(os.Getenv("TELEGRAM_TOKEN"), opts...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
}

func worker(ctx context.Context, b *bot.Bot) {
	for {
		if len(scheduledPosts) == 0 {
			time.Sleep(10 * time.Second)
			continue
		}

		now := time.Now()
		post := &scheduledPosts[0]
		if (post.SendAt.Before(now)) {
			b.CopyMessage(ctx, &bot.CopyMessageParams{
				ChatID: post.ChatID,
				FromChatID: post.FromChatID,
				MessageID: post.Message_id,
			})
			scheduledPosts = scheduledPosts[1:]
		}
	}
}