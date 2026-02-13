package shopifywebhook

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool_ProcessesEvents(t *testing.T) {
	var count atomic.Int32

	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error {
		count.Add(1)
		return nil
	})

	pool := NewWorkerPool(2, 100)

	for range 10 {
		pool.Submit(Event{
			Metadata: Metadata{Topic: TopicOrdersCreate},
			RawBody:  []byte(`{}`),
		}, router)
	}

	if err := pool.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}

	if got := count.Load(); got != 10 {
		t.Fatalf("expected 10 events processed, got %d", got)
	}
}

func TestWorkerPool_QueueFull(t *testing.T) {
	var dropped atomic.Int32

	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error {
		time.Sleep(50 * time.Millisecond) // Simulate slow processing.
		return nil
	})

	pool := NewWorkerPool(1, 1, WithPoolErrorHandler(func(event Event, err error) {
		if errors.Is(err, ErrQueueFull) {
			dropped.Add(1)
		}
	}))

	// Fill the queue and the single worker.
	for range 20 {
		pool.Submit(Event{
			Metadata: Metadata{Topic: TopicOrdersCreate},
			RawBody:  []byte(`{}`),
		}, router)
	}

	_ = pool.Shutdown(context.Background())

	if got := dropped.Load(); got == 0 {
		t.Fatal("expected at least one dropped event when queue is full")
	}
}

func TestWorkerPool_ShutdownTimeout(t *testing.T) {
	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error {
		time.Sleep(5 * time.Second)
		return nil
	})

	pool := NewWorkerPool(1, 10)
	pool.Submit(Event{
		Metadata: Metadata{Topic: TopicOrdersCreate},
		RawBody:  []byte(`{}`),
	}, router)

	// Give worker time to pick up the event.
	time.Sleep(10 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := pool.Shutdown(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got: %v", err)
	}
}

func TestWorkerPool_RetriesOnFailure(t *testing.T) {
	var attempts atomic.Int32

	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error {
		n := attempts.Add(1)
		if n < 3 {
			return errors.New("transient failure")
		}
		return nil // Succeeds on 3rd attempt.
	})

	pool := NewWorkerPool(1, 100,
		WithMaxRetries(3),
		WithRetryBaseDelay(10*time.Millisecond),
	)

	pool.Submit(Event{
		Metadata: Metadata{Topic: TopicOrdersCreate},
		RawBody:  []byte(`{}`),
	}, router)

	_ = pool.Shutdown(context.Background())

	if got := attempts.Load(); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestWorkerPool_ExhaustsRetries(t *testing.T) {
	var finalErr atomic.Value

	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error {
		return errors.New("permanent failure")
	})

	pool := NewWorkerPool(1, 100,
		WithMaxRetries(2),
		WithRetryBaseDelay(10*time.Millisecond),
		WithPoolErrorHandler(func(event Event, err error) {
			if !errors.Is(err, ErrQueueFull) {
				finalErr.Store(err)
			}
		}),
	)

	pool.Submit(Event{
		Metadata: Metadata{Topic: TopicOrdersCreate},
		RawBody:  []byte(`{}`),
	}, router)

	_ = pool.Shutdown(context.Background())

	stored := finalErr.Load()
	if stored == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if stored.(error).Error() != "permanent failure" {
		t.Fatalf("unexpected error: %v", stored)
	}
}

func TestWorkerPool_NoRetriesByDefault(t *testing.T) {
	var errorCount atomic.Int32

	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error {
		return errors.New("fail")
	})

	pool := NewWorkerPool(1, 100,
		WithPoolErrorHandler(func(event Event, err error) {
			errorCount.Add(1)
		}),
	)

	pool.Submit(Event{
		Metadata: Metadata{Topic: TopicOrdersCreate},
		RawBody:  []byte(`{}`),
	}, router)

	_ = pool.Shutdown(context.Background())

	if got := errorCount.Load(); got != 1 {
		t.Fatalf("expected 1 error (no retries), got %d", got)
	}
}

func TestWorkerPool_HandlerErrors(t *testing.T) {
	var capturedErr atomic.Value

	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error {
		return errors.New("handler failed")
	})

	pool := NewWorkerPool(1, 10, WithPoolErrorHandler(func(event Event, err error) {
		capturedErr.Store(err)
	}))

	pool.Submit(Event{
		Metadata: Metadata{Topic: TopicOrdersCreate},
		RawBody:  []byte(`{}`),
	}, router)

	_ = pool.Shutdown(context.Background())

	stored := capturedErr.Load()
	if stored == nil {
		t.Fatal("expected error handler to be called")
	}
	if stored.(error).Error() != "handler failed" {
		t.Fatalf("unexpected error: %v", stored)
	}
}
