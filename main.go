package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type ScheduledPost struct {
	ChatID     int64
	FromChatID int64
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
	go worker(ctx, b)

	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	newPost := ScheduledPost{
		ChatID: update.Message.Chat.ID,
		FromChatID: update.Message.From.ID,
		Message_id: update.Message.ID,
	}
	if (len(scheduledPosts) == 0) {
		newPost.SendAt = time.Now().Add(time.Hour)
	} else {
		delay := scheduledPosts[len(scheduledPosts)-1].SendAt.Add(time.Hour)
		newPost.SendAt = delay.Add(time.Duration(rand.Intn(20) - 10) * time.Minute)
	}
	scheduledPosts = append(scheduledPosts, newPost)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: "Post scheduled",
		ReplyParameters: &models.ReplyParameters{
			MessageID: update.Message.ID,
		},
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