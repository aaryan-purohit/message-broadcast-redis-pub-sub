package processor

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"main/internal/dispatcher"
	"main/internal/events"
)

// testHandler records handled messages deterministically
type testHandler struct {
	mu     sync.Mutex
	count  int
	seenCh chan struct{}
	err    error
}

func newTestHandler(expected int) *testHandler {
	return &testHandler{
		seenCh: make(chan struct{}, expected),
	}
}

func (h *testHandler) Handle(ctx context.Context, msg events.Message) error {
	h.mu.Lock()
	h.count++
	h.mu.Unlock()

	h.seenCh <- struct{}{}
	return h.err
}

func waitForMessages(t *testing.T, h *testHandler, n int) {
	t.Helper()

	timeout := time.After(1 * time.Second)
	for i := 0; i < n; i++ {
		select {
		case <-h.seenCh:
		case <-timeout:
			t.Fatalf("timeout waiting for %d messages, got %d", n, i)
		}
	}
}

func TestNew(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	proc := New(d, logger, 2, 10)

	if proc == nil {
		t.Fatal("expected processor to be non-nil")
	}
	if proc.dispatcher != d {
		t.Fatal("expected dispatcher to be set")
	}
}

func TestProcessor_Submit_DispatchesMessage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	handler := newTestHandler(1)
	d.Register("test.event", handler)

	proc := New(d, logger, 1, 10)
	defer proc.Stop()

	msg := events.Message{
		ID:   "msg-1",
		Type: "test.event",
	}

	if err := proc.Submit(msg); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	waitForMessages(t, handler, 1)

	metrics := proc.GetMetrics()
	if metrics["processed"] != 1 {
		t.Fatalf("expected 1 processed, got %d", metrics["processed"])
	}
}

func TestProcessor_Submit_HandlerErrorStillCountsProcessed(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	handler := newTestHandler(1)
	handler.err = errors.New("handler failed")
	d.Register("test.event", handler)

	proc := New(d, logger, 1, 10)
	defer proc.Stop()

	_ = proc.Submit(events.Message{ID: "1", Type: "test.event"})

	waitForMessages(t, handler, 1)

	metrics := proc.GetMetrics()
	if metrics["processed"] != 1 {
		t.Fatalf("expected processed=1, got %d", metrics["processed"])
	}
}

func TestProcessor_Submit_QueueFull(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	proc := New(d, logger, 0, 1) // no workers, queue size 1
	defer proc.Stop()

	if err := proc.Submit(events.Message{ID: "1", Type: "x"}); err != nil {
		t.Fatalf("first submit failed: %v", err)
	}

	err := proc.Submit(events.Message{ID: "2", Type: "x"})
	if err != ErrQueueFull {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}

	metrics := proc.GetMetrics()
	if metrics["dropped"] != 1 {
		t.Fatalf("expected dropped=1, got %d", metrics["dropped"])
	}
	if metrics["queued"] != 1 {
		t.Fatalf("expected queued=1, got %d", metrics["queued"])
	}
}

func TestProcessor_ConcurrentWorkers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	handler := newTestHandler(10)
	d.Register("test.event", handler)

	proc := New(d, logger, 3, 10)
	defer proc.Stop()

	for i := 0; i < 10; i++ {
		err := proc.Submit(events.Message{
			ID:   "msg",
			Type: "test.event",
		})
		if err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}

	waitForMessages(t, handler, 10)

	metrics := proc.GetMetrics()
	if metrics["processed"] != 10 {
		t.Fatalf("expected processed=10, got %d", metrics["processed"])
	}
}

func TestProcessor_Submit_AfterStopFails(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	proc := New(d, logger, 1, 10)
	proc.Stop()

	err := proc.Submit(events.Message{ID: "x", Type: "x"})
	if err == nil {
		t.Fatal("expected error when submitting after Stop")
	}
}

func TestProcessor_Stop_DrainsQueue(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	handler := newTestHandler(3)
	d.Register("test.event", handler)

	proc := New(d, logger, 1, 10)

	for i := 0; i < 3; i++ {
		_ = proc.Submit(events.Message{ID: "x", Type: "test.event"})
	}

	proc.Stop()

	waitForMessages(t, handler, 3)

	metrics := proc.GetMetrics()
	if metrics["processed"] != 3 {
		t.Fatalf("expected processed=3, got %d", metrics["processed"])
	}
}

func TestProcessor_GetMetrics_AlwaysPresent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	proc := New(d, logger, 1, 10)
	defer proc.Stop()

	metrics := proc.GetMetrics()

	keys := []string{"processed", "dropped", "queued"}
	for _, k := range keys {
		if _, ok := metrics[k]; !ok {
			t.Fatalf("expected metric %q to exist", k)
		}
	}
}
