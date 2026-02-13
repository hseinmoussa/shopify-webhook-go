package shopifywebhook

import (
	"context"
	"sync"
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
type WorkerPool struct {
	queue   chan work
	wg      sync.WaitGroup
	onError ErrorHandlerFunc
}

type work struct {
	event  Event
	router *Router
}

// NewWorkerPool creates a pool with the specified number of workers and queue capacity.
//
// If the queue is full when Submit is called, the event is dropped and
// onError is called with ErrQueueFull. This is intentional: blocking the
// HTTP goroutine would cause Shopify to time out and retry anyway.
//
// Typical production values: workers=10, queueSize=1000.
func NewWorkerPool(workers, queueSize int, opts ...WorkerPoolOption) *WorkerPool {
	cfg := &workerPoolConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	wp := &WorkerPool{
		queue:   make(chan work, queueSize),
		onError: cfg.onError,
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
		if err := w.router.Dispatch(w.event); err != nil {
			if wp.onError != nil {
				wp.onError(w.event, err)
			}
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
	onError ErrorHandlerFunc
}

// WithPoolErrorHandler sets the error handler for processing errors
// and queue-full events.
func WithPoolErrorHandler(fn ErrorHandlerFunc) WorkerPoolOption {
	return func(c *workerPoolConfig) {
		c.onError = fn
	}
}
