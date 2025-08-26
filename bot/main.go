package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type ScheduledPost struct {
	ChatID     int64
	FromChatID int64
	Message_id int
	SendAt     time.Time
}

var scheduledPosts []ScheduledPost
var dbpool *pgxpool.Pool

func main() {
	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	var err error
	dbpool, err = pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(os.Getenv("BOT_TOKEN"), opts...)
	if err != nil {
		panic(err)
	}
	go worker(ctx, b)

	b.Start(ctx)
}

func savePost(ctx context.Context, dbpool *pgxpool.Pool, post ScheduledPost) error {
	query := "INSERT INTO scheduled_posts (chat_id, from_chat_id, message_id, send_at) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING"
	log.Printf("Saving post: %v\n", post)
	_, err := dbpool.Exec(ctx, query, post.ChatID, post.FromChatID, post.Message_id, post.SendAt)
	return err
}

func getPostsToSend(ctx context.Context, dbpool *pgxpool.Pool) ([]ScheduledPost, error) {
	var posts []ScheduledPost
	currentTime := time.Now()
	query := "SELECT chat_id, from_chat_id, message_id, send_at FROM scheduled_posts WHERE send_at <= $1 ORDER BY send_at ASC"
	rows, err := dbpool.Query(ctx, query, currentTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var post ScheduledPost
		err := rows.Scan(
			&post.ChatID,
			&post.FromChatID,
			&post.Message_id,
			&post.SendAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	query = "DELETE FROM scheduled_posts WHERE send_at <= $1"
	_, err = dbpool.Exec(ctx, query, currentTime)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func sendPosts(ctx context.Context, b *bot.Bot, posts []ScheduledPost) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)
	for _, post := range posts {
		post := post
		g.Go(func() error {
			b.CopyMessage(ctx, &bot.CopyMessageParams{
				ChatID:		 post.ChatID,
				FromChatID: post.FromChatID,
				MessageID:	post.Message_id,
			})
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Printf("Error sending posts: %v\n", err)
	}
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("update %+v\n", update.MyChatMember)
	if update.Message == nil {
		return
	}
	newPost := ScheduledPost{
		ChatID:     update.Message.Chat.ID,
		FromChatID: update.Message.From.ID,
		Message_id: update.Message.ID,
	}
	// send_delay := time.Hour + time.Duration(rand.Intn(10))*time.Minute
	send_delay := time.Minute;
	if len(scheduledPosts) == 0 {
		newPost.SendAt = time.Now().Truncate(time.Hour).Add(send_delay)
	} else {
		newPost.SendAt = scheduledPosts[len(scheduledPosts)-1].SendAt.Truncate(time.Hour).Add(send_delay)
	}
	scheduledPosts = append(scheduledPosts, newPost)
	savePost(ctx, dbpool, newPost)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Post scheduled to %s", newPost.SendAt.Format(time.RFC3339)),
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
		posts, err := getPostsToSend(ctx, dbpool)
		if err != nil {
			log.Printf("Error getting posts to send: %v\n", err)
			continue
		}
		sendPosts(ctx, b, posts)
	}
}
