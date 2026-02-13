package shopifywebhook

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// AsyncProcessor submits events for background processing.
// Implement this interface to use a custom queue (e.g., SQS, Kafka, Redis).
type AsyncProcessor interface {
	// Submit enqueues an event for processing. Must not block.
	Submit(event Event, router *Router)

	// Shutdown gracefully waits for pending events to complete.
	// The context can set a deadline for the shutdown.
	Shutdown(ctx context.Context) error
}

// WorkerPool is a channel-based AsyncProcessor with a fixed number of workers.
//
// By default, failed events are reported to the error handler and discarded.
// Use WithMaxRetries to enable automatic retries with exponential backoff.
type WorkerPool struct {
	queue      chan work
	wg         sync.WaitGroup
	onError    ErrorHandlerFunc
	maxRetries int
	baseDelay  time.Duration
	closing    atomic.Bool
}

type work struct {
	event   Event
	router  *Router
	attempt int
}

// NewWorkerPool creates a pool with the specified number of workers and queue capacity.
//
// If the queue is full when Submit is called, the event is dropped and
// onError is called with ErrQueueFull. This is intentional: blocking the
// HTTP goroutine would cause Shopify to time out and retry anyway.
//
// Typical production values: workers=10, queueSize=1000.
func NewWorkerPool(workers, queueSize int, opts ...WorkerPoolOption) *WorkerPool {
	cfg := &workerPoolConfig{
		baseDelay: 500 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	wp := &WorkerPool{
		queue:      make(chan work, queueSize),
		onError:    cfg.onError,
		maxRetries: cfg.maxRetries,
		baseDelay:  cfg.baseDelay,
	}

	wp.wg.Add(workers)
	for range workers {
		go wp.worker()
	}

	return wp
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for w := range wp.queue {
		wp.processWithRetry(w)
	}
}

func (wp *WorkerPool) processWithRetry(w work) {
	for attempt := range wp.maxRetries + 1 {
		err := w.router.Dispatch(w.event)
		if err == nil {
			return
		}

		if attempt < wp.maxRetries {
			// Exponential backoff: 500ms, 1s, 2s, 4s, ...
			delay := wp.baseDelay * time.Duration(math.Pow(2, float64(attempt)))
			time.Sleep(delay)
			continue
		}

		// Max retries exhausted (or no retries configured).
		if wp.onError != nil {
			wp.onError(w.event, err)
		}
	}
}

// Submit enqueues an event for background processing.
// Non-blocking: drops the event if the queue is full.
func (wp *WorkerPool) Submit(event Event, router *Router) {
	select {
	case wp.queue <- work{event: event, router: router}:
	default:
		if wp.onError != nil {
			wp.onError(event, ErrQueueFull)
		}
	}
}

// Shutdown closes the queue and waits for all workers to finish processing.
// Respects the context deadline.
func (wp *WorkerPool) Shutdown(ctx context.Context) error {
	wp.closing.Store(true)
	close(wp.queue)
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WorkerPoolOption configures a WorkerPool.
type WorkerPoolOption func(*workerPoolConfig)

type workerPoolConfig struct {
	onError    ErrorHandlerFunc
	maxRetries int
	baseDelay  time.Duration
}

// WithPoolErrorHandler sets the error handler for processing errors
// and queue-full events.
func WithPoolErrorHandler(fn ErrorHandlerFunc) WorkerPoolOption {
	return func(c *workerPoolConfig) {
		c.onError = fn
	}
}

// WithMaxRetries enables automatic retries with exponential backoff.
// Failed events are re-enqueued up to maxRetries times before being
// reported to the error handler and discarded.
//
// Backoff schedule (with default 500ms base delay):
//
//	Attempt 1: 500ms
//	Attempt 2: 1s
//	Attempt 3: 2s
func WithMaxRetries(maxRetries int) WorkerPoolOption {
	return func(c *workerPoolConfig) {
		c.maxRetries = maxRetries
	}
}

// WithRetryBaseDelay sets the base delay for exponential backoff.
// Default: 500ms. The delay doubles on each retry attempt.
func WithRetryBaseDelay(d time.Duration) WorkerPoolOption {
	return func(c *workerPoolConfig) {
		c.baseDelay = d
	}
}
