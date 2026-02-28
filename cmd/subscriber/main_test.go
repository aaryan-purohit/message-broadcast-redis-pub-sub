package main

import (
	"os"
	"testing"
)

func TestGetEnv_WithEnvironmentVariable(t *testing.T) {
	key := "TEST_KEY_SUB"
	expected := "test_value"

	os.Setenv(key, expected)
	defer os.Unsetenv(key)

	result := getEnv(key, "fallback")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetEnv_WithFallback(t *testing.T) {
	key := "NON_EXISTENT_KEY_SUB"
	fallback := "fallback_value"

	os.Unsetenv(key)

	result := getEnv(key, fallback)
	if result != fallback {
		t.Errorf("expected %q, got %q", fallback, result)
	}
}

func TestGetEnv_WithEmptyEnvironmentVariable(t *testing.T) {
	key := "EMPTY_KEY_SUB"
	fallback := "fallback_value"

	os.Setenv(key, "")
	defer os.Unsetenv(key)

	result := getEnv(key, fallback)
	if result != fallback {
		t.Errorf("expected %q when env var is empty, got %q", fallback, result)
	}
}

func TestGetEnv_WithDifferentValues(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback string
		expected string
	}{
		{"empty string value", "EMPTY_SUB", "", "default", "default"},
		{"non-empty value", "REDIS_ADDR_SUB", "localhost:6379", "127.0.0.1:6379", "localhost:6379"},
		{"missing key", "MISSING_SUB", "", "default-value", "default-value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.fallback)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
