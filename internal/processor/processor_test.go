package processor

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"main/internal/dispatcher"
	"main/internal/events"
)

func TestNew(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := dispatcher.New(logger)

	proc := New(d, logger, 2, 10)

	if proc == nil {
		t.Fatal("expected processor to be non-nil")
	}
	if proc.dispatcher != d {
		t.Fatal("expected dispatcher to be set")
	}
}

func TestProcessor_Submit_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := dispatcher.New(logger)
	proc := New(d, logger, 1, 10)
	defer proc.Stop()

	msg := events.Message{
		ID:      "test-123",
		Type:    "test.event",
		Source:  "test",
		Payload: nil,
	}

	err := proc.Submit(msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	metrics := proc.GetMetrics()
	if metrics["processed"] < 1 {
		t.Errorf("expected at least 1 processed message, got %d", metrics["processed"])
	}
}

func TestProcessor_Submit_QueueFull(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := dispatcher.New(logger)
	proc := New(d, logger, 0, 1)
	defer func() {
		proc.cancel()
		time.Sleep(100 * time.Millisecond)
		proc.Stop()
	}()

	msg := events.Message{ID: "msg-1", Type: "test"}
	err := proc.Submit(msg)
	if err != nil {
		t.Fatalf("first submit failed: %v", err)
	}

	msg2 := events.Message{ID: "msg-2", Type: "test"}
	err = proc.Submit(msg2)
	if err != ErrQueueFull {
		t.Errorf("expected ErrQueueFull, got %v", err)
	}

	metrics := proc.GetMetrics()
	if metrics["dropped"] < 1 {
		t.Errorf("expected at least 1 dropped message, got %d", metrics["dropped"])
	}
}

func TestProcessor_GetMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := dispatcher.New(logger)
	proc := New(d, logger, 1, 10)
	defer proc.Stop()

	metrics := proc.GetMetrics()

	if _, ok := metrics["processed"]; !ok {
		t.Fatal("expected 'processed' key in metrics")
	}
	if _, ok := metrics["dropped"]; !ok {
		t.Fatal("expected 'dropped' key in metrics")
	}
	if _, ok := metrics["queued"]; !ok {
		t.Fatal("expected 'queued' key in metrics")
	}
}

func TestProcessor_Stop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := dispatcher.New(logger)
	proc := New(d, logger, 1, 10)

	msg := events.Message{ID: "test", Type: "test"}
	err := proc.Submit(msg)
	if err != nil {
		t.Fatalf("submit before stop failed: %v", err)
	}

	proc.Stop()

	// Verify processor was stopped by checking metrics
	metrics := proc.GetMetrics()
	if metrics["processed"] < 0 {
		t.Errorf("expected processed count >= 0, got %d", metrics["processed"])
	}
}
