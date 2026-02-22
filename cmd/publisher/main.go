package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"main/internal/events"
	"main/internal/redisclient"
	"os"
	"time"

	"github.com/google/uuid"
)

func main() {
	// -------- Config --------
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	channel := getEnv("CHANNEL_NAME", "broadcast.events")
	source := getEnv("SERVER_ID", "publisher")

	// -------- Logger --------
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	).With("component", "publisher", "source", source)

	slog.SetDefault(logger)

	// Redis
	rdb, err := redisclient.New(redisAddr, 0)
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		event := events.Message{
			ID:        uuid.NewString(),
			Type:      "demo.message",
			Source:    source,
			Timestamp: time.Now(),
			Payload: map[string]any{
				"counter": i,
				"text":    "hello from publisher",
			},
		}
		data, err := json.Marshal(event)
		if err != nil {
			logger.Error("failed to marshall event", "error", err)
			continue
		}
		if err := rdb.Publish(ctx, channel, data).Err(); err != nil {
			logger.Error("failed to publis message", "error", err)
			continue
		}
		logger.Info("message published", "event_id", event.ID, "type", event.Type)
		time.Sleep(2 * time.Second)
	}

}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
