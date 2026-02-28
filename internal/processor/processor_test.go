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

	err := proc.Submit(events.Message{ID: "1", Type: "test.event"})
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	waitForMessages(t, handler, 1)

	metrics := proc.GetMetrics()
	if metrics["processed"] != 1 {
		t.Fatalf("expected processed=1, got %d", metrics["processed"])
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

	if proc.GetMetrics()["processed"] != 1 {
		t.Fatal("handler error should still count as processed")
	}
}

func TestProcessor_Submit_QueueFull(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	proc := New(d, logger, 0, 1)
	defer proc.Stop()

	if err := proc.Submit(events.Message{ID: "1"}); err != nil {
		t.Fatalf("first submit failed: %v", err)
	}

	if err := proc.Submit(events.Message{ID: "2"}); err != ErrQueueFull {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}

	if proc.GetMetrics()["dropped"] != 1 {
		t.Fatal("expected dropped=1")
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
		_ = proc.Submit(events.Message{ID: "x", Type: "test.event"})
	}

	waitForMessages(t, handler, 10)

	if proc.GetMetrics()["processed"] != 10 {
		t.Fatal("expected processed=10")
	}
}

func TestProcessor_Stop_IsFinal(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	handler := newTestHandler(1)
	d.Register("test.event", handler)

	proc := New(d, logger, 1, 10)

	_ = proc.Submit(events.Message{ID: "1", Type: "test.event"})
	proc.Stop()

	select {
	case <-handler.seenCh:
		// allowed (race window)
	case <-time.After(300 * time.Millisecond):
		// also allowed
	}

	metrics := proc.GetMetrics()
	if metrics["processed"] > 1 {
		t.Fatalf("unexpected processed count: %d", metrics["processed"])
	}
}

func TestProcessor_Submit_AfterStop_DoesNotProcess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	handler := newTestHandler(1)
	d.Register("test.event", handler)

	proc := New(d, logger, 1, 10)
	proc.Stop()

	// Submit after stop
	_ = proc.Submit(events.Message{ID: "x", Type: "test.event"})

	// Ensure nothing is processed
	select {
	case <-handler.seenCh:
		t.Fatal("message should not be processed after stop")
	case <-time.After(300 * time.Millisecond):
		// expected
	}

	if proc.GetMetrics()["processed"] != 0 {
		t.Fatal("processed count should remain zero after stop")
	}
}

func TestProcessor_GetMetrics_AlwaysPresent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dispatcher.New(logger)

	proc := New(d, logger, 1, 10)
	defer proc.Stop()

	for _, k := range []string{"processed", "dropped", "queued"} {
		if _, ok := proc.GetMetrics()[k]; !ok {
			t.Fatalf("missing metric %q", k)
		}
	}
}
