package processor

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aaryan-purohit/message-broadcast-redis-pub-sub/internal/dispatcher"
	"github.com/aaryan-purohit/message-broadcast-redis-pub-sub/internal/events"
)

func TestNewProcessor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatcher.New(logger)

	p := New(dispatcher, logger, 2, 10)
	if p == nil {
		t.Fatal("expected processor to be created, got nil")
	}
	if p.dispatcher != dispatcher {
		t.Fatal("expected dispatcher to be set correctly")
	}
	if p.logger != logger {
		t.Fatal("expected logger to be set correctly")
	}
	if len(p.queue) != 0 {
		t.Fatal("expected queue to be empty on initialization")
	}
	if p.maxRetries != 3 {
		t.Fatalf("expected maxRetries to be 3, got %d", p.maxRetries)
	}
	if p.retryDelay != 100*time.Millisecond {
		t.Fatalf("expected retryDelay to be 100ms, got %s", p.retryDelay)
	}

}

func TestProcessorSubmit(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatcher.New(logger)
	p := New(dispatcher, logger, 1, 2)

	// Submit a message and expect it to be accepted
	err := p.Submit(events.Message{ID: "1", Payload: "test"})
	if err != nil {
		t.Fatalf("expected message to be accepted, got error: %v", err)
	}
}

func TestProcessorSubmitQueueFull(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatcher.New(logger)
	p := New(dispatcher, logger, 1, 1)

	// Fill the queue
	err := p.Submit(events.Message{ID: "1", Payload: "test"})
	if err != nil {
		t.Fatalf("expected message to be accepted, got error: %v", err)
	}

	// Submit another message and expect it to be rejected due to full queue
	err = p.Submit(events.Message{ID: "2", Payload: "test"})
	if err != ErrQueueFull {
		t.Fatalf("expected ErrQueueFull, got: %v", err)
	}
}

func TestProcessorSubmitAfterShutdown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatcher.New(logger)
	p := New(dispatcher, logger, 1, 10)

	// Shutdown the processor
	p.Stop()

	// Submit a message and expect it to be rejected due to shutdown
	err := p.Submit(events.Message{ID: "1", Payload: "test"})
	if err == nil {
		t.Fatal("expected error when submitting after shutdown, got nil")
	}
}

func TestProcessorGetMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatcher.New(logger)
	p := New(dispatcher, logger, 1, 10)

	metrics := p.GetMetrics()
	if metrics["processed"] != 0 {
		t.Fatalf("expected processed to be 0, got %d", metrics["processed"])
	}
	if metrics["dropped"] != 0 {
		t.Fatalf("expected dropped to be 0, got %d", metrics["dropped"])
	}
	if metrics["queued"] != 0 {
		t.Fatalf("expected queued to be 0, got %d", metrics["queued"])
	}
}

// --- worker tests -------------------------------------------------------

// fakeHandler keeps track of how many times it was called and can fail
// a configurable number of times before succeeding. This allows us to
// exercise the retry loop in the processor.worker implementation.

type fakeHandler struct {
	calls    atomic.Int32
	failures int // number of times to return an error before success
}

func (h *fakeHandler) Handle(ctx context.Context, event events.Message) error {
	count := h.calls.Add(1)
	if int(count) <= h.failures {
		return errors.New("simulated failure")
	}
	return nil
}

func TestWorkerSuccessfulDispatch(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := dispatcher.New(logger)

	h := &fakeHandler{failures: 0}
	d.Register("test", h)

	p := New(d, logger, 1, 1)

	// send a single message and allow the worker to process it
	if err := p.Submit(events.Message{ID: "1", Type: "test"}); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	// wait for the handler to be invoked or timeout
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) && h.calls.Load() == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	// stop the processor which will wait for the worker to finish
	p.Stop()

	if h.calls.Load() != 1 {
		t.Fatalf("expected handler to be called once, got %d", h.calls.Load())
	}
	if p.processed.Load() != 1 {
		t.Fatalf("expected processed count to be 1, got %d", p.processed.Load())
	}
}

func TestWorkerRetriesOnFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := dispatcher.New(logger)

	// configure the handler to fail twice before succeeding; our
	// processor is configured to retry up to 3 times by default.
	h := &fakeHandler{failures: 2}
	d.Register("test", h)

	p := New(d, logger, 1, 1)

	if err := p.Submit(events.Message{ID: "2", Type: "test"}); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	// wait for retries to complete (handler.calls should reach 3)
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) && h.calls.Load() < int32(3) {
		time.Sleep(10 * time.Millisecond)
	}

	// stop the processor, waiting for the worker and retry loop
	p.Stop()

	// the handler should have been invoked 3 times (2 failures + 1 success)
	if h.calls.Load() != int32(3) {
		t.Fatalf("expected handler to be called three times, got %d", h.calls.Load())
	}
	// even though there were failures, processor.worker increments processed
	// after the call, so the metric should be 1.
	if p.processed.Load() != 1 {
		t.Fatalf("expected processed count to be 1 after retries, got %d", p.processed.Load())
	}
}

func TestWorkerExceedsMaxRetries(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := dispatcher.New(logger)

	// configure the handler to fail more times than the processor's maxRetries
	h := &fakeHandler{failures: 5}
	d.Register("test", h)
	p := New(d, logger, 1, 1)

	if err := p.Submit(events.Message{ID: "3", Type: "test"}); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	// wait for retries to complete (handler.calls should reach 4: 3 failures + 1 final attempt)
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) && h.calls.Load() < int32(4) {
		time.Sleep(10 * time.Millisecond)
	}

	// stop the processor, waiting for the worker and retry loop
	p.Stop()

	// the handler should have been invoked 4 times (3 failures + 1 final attempt)
	if h.calls.Load() != int32(4) {
		t.Fatalf("expected handler to be called four times, got %d", h.calls.Load())
	}
	// even though all attempts failed, processor.worker increments processed
	// after the call, so the metric should be 1.
	if p.processed.Load() != 1 {
		t.Fatalf("expected processed count to be 1 after exceeding retries, got %d", p.processed.Load())
	}
}
