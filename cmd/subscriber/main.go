package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"log/slog"

	"main/internal/dispatcher"
	"main/internal/handlers"
	"main/internal/processor"
	"main/internal/redisclient"
)

func main() {
	// -------- Config --------
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	channel := getEnv("CHANNEL_NAME", "broadcast.events")
	serverID := getEnv("SERVER_ID", "unknown-server")

	// -------- Logger --------
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	).With("server_id", serverID, "component", "subscriber")

	slog.SetDefault(logger)

	// -------- Redis --------
	rdb, err := redisclient.New(redisAddr, 0)
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	d := dispatcher.New(logger)
	d.Register("demo.message", handlers.NewDemoMessageHandler(logger))

	p := processor.New(d, logger, 4, 100)

	sub := redisclient.NewSubscriber(rdb, channel, p, logger)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("shutdown signal received")
		p.Stop()
		os.Exit(0)
	}()

	if err := sub.Start(ctx); err != nil {
		logger.Error("subscriber stopped", "error", err)
		p.Stop()
	}

}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
