// Async example: worker pool + deduplication.
// Responds 200 immediately, processes webhooks in background workers.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	sw "github.com/hseinmoussa/shopify-webhook-go"
)

func main() {
	secret := os.Getenv("SHOPIFY_WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("SHOPIFY_WEBHOOK_SECRET is required")
	}

	router := sw.NewRouter(
		sw.WithErrorHandler(func(event sw.Event, err error) {
			log.Printf("ERROR [%s] event=%s: %v",
				event.Metadata.Topic, event.Metadata.EventID, err)
		}),
	)

	router.Handle(sw.TopicOrdersCreate, func(event sw.Event) error {
		var order sw.Order
		if err := event.Unmarshal(&order); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		// Simulate slow processing (DB writes, external APIs, etc.)
		// This would timeout Shopify without async processing.
		time.Sleep(2 * time.Second)

		log.Printf("Processed order #%d", order.OrderNumber)
		return nil
	})

	// Worker pool: 10 workers, queue of 1000.
	pool := sw.NewWorkerPool(10, 1000,
		sw.WithPoolErrorHandler(func(event sw.Event, err error) {
			log.Printf("WORKER ERROR [%s]: %v", event.Metadata.Topic, err)
		}),
	)

	// Dedup store: skip webhooks already processed in the last 24h.
	store := sw.NewMemoryStore(24 * time.Hour)

	handler := sw.Handler(secret, router,
		sw.WithAsyncProcessor(pool),
		sw.WithIdempotencyStore(store),
	)

	http.Handle("/webhooks", handler)

	srv := &http.Server{Addr: ":8080"}

	// Graceful shutdown.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Stop accepting new requests.
		_ = srv.Shutdown(ctx)

		// Drain the worker pool.
		if err := pool.Shutdown(ctx); err != nil {
			log.Printf("Worker pool shutdown error: %v", err)
		}

		store.Close()
		log.Println("Shutdown complete")
	}()

	log.Println("Listening on :8080 (async mode, 10 workers)")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
