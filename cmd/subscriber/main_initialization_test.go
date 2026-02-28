package main

import (
	"log/slog"
	"os"
	"testing"

	"main/internal/dispatcher"
	"main/internal/handlers"
	"main/internal/processor"
)

func TestInit_DispatcherWithHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := dispatcher.New(logger)

	handler := handlers.NewDemoMessageHandler(logger)
	d.Register("demo.message", handler)

	if d == nil {
		t.Fatal("expected dispatcher to be initialized")
	}
}

func TestInit_ProcessorWithDispatcher(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := dispatcher.New(logger)

	p := processor.New(d, logger, 4, 100)

	if p == nil {
		t.Fatal("expected processor to be initialized")
	}
}

func TestInit_ConfigFromEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envVal   string
		fallback string
		expected string
	}{
		{"REDIS_ADDR override", "REDIS_ADDR", "custom:6380", "localhost:6379", "custom:6380"},
		{"CHANNEL_NAME override", "CHANNEL_NAME", "custom.events", "broadcast.events", "custom.events"},
		{"SERVER_ID override", "SERVER_ID", "custom-server", "unknown-server", "custom-server"},
		{"REDIS_ADDR default", "REDIS_ADDR_UNUSED", "", "localhost:6379", "localhost:6379"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				os.Setenv(tt.envKey, tt.envVal)
				defer os.Unsetenv(tt.envKey)
			} else {
				os.Unsetenv(tt.envKey)
			}

			var result string
			if tt.envKey == "REDIS_ADDR_UNUSED" {
				result = getEnv("REDIS_ADDR", tt.fallback)
			} else {
				result = getEnv(tt.envKey, tt.fallback)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInit_LoggerCreation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if logger == nil {
		t.Fatal("expected logger to be created")
	}
}

func TestInit_ComponentsNotNil(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	d := dispatcher.New(logger)
	if d == nil {
		t.Error("expected dispatcher to be non-nil")
	}

	handler := handlers.NewDemoMessageHandler(logger)
	if handler == nil {
		t.Error("expected handler to be non-nil")
	}

	p := processor.New(d, logger, 4, 100)
	if p == nil {
		t.Error("expected processor to be non-nil")
	}

	defer p.Stop()
}

func TestInit_DefaultConfigValues(t *testing.T) {
	os.Unsetenv("REDIS_ADDR")
	os.Unsetenv("CHANNEL_NAME")
	os.Unsetenv("SERVER_ID")

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	channel := getEnv("CHANNEL_NAME", "broadcast.events")
	serverID := getEnv("SERVER_ID", "unknown-server")

	if redisAddr != "localhost:6379" {
		t.Errorf("expected default REDIS_ADDR, got %q", redisAddr)
	}
	if channel != "broadcast.events" {
		t.Errorf("expected default CHANNEL_NAME, got %q", channel)
	}
	if serverID != "unknown-server" {
		t.Errorf("expected default SERVER_ID, got %q", serverID)
	}
}
