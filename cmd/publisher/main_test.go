package main

import (
	"os"
	"testing"
)

func TestGetEnv_ReturnsEnvValue(t *testing.T) {
	key := "TEST_ENV_KEY"
	expected := "actual_value"

	os.Setenv(key, expected)
	defer os.Unsetenv(key)

	result := getEnv(key, "fallback_value")

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestGetEnv_ReturnsFallbackWhenNotSet(t *testing.T) {
	key := "TEST_ENV_KEY_NOT_SET"
	fallback := "fallback_value"

	os.Unsetenv(key)

	result := getEnv(key, fallback)

	if result != fallback {
		t.Errorf("expected fallback %s, got %s", fallback, result)
	}
}
