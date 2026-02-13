package shopifywebhook

import (
	"context"
	"sync"
	"time"
)

// IdempotencyStore tracks processed webhook event IDs for deduplication.
//
// Implement this interface for your storage backend:
//   - Redis: use SETNX with TTL
//   - PostgreSQL: INSERT ... ON CONFLICT DO NOTHING
//   - DynamoDB: conditional PutItem
type IdempotencyStore interface {
	// Exists returns true if the event ID has already been processed.
	Exists(ctx context.Context, eventID string) (bool, error)

	// Store marks an event ID as processed.
	Store(ctx context.Context, eventID string) error
}

// MemoryStore is an in-memory IdempotencyStore suitable for
// single-instance deployments.
//
// Entries automatically expire after the configured TTL to prevent
// unbounded memory growth. A background goroutine evicts expired
// entries periodically.
type MemoryStore struct {
	mu      sync.RWMutex
	entries map[string]time.Time
	ttl     time.Duration
	done    chan struct{}
}

// NewMemoryStore creates a MemoryStore with the given TTL.
//
// Typical TTL: 24 hours. Shopify retries for up to 48 hours,
// but 24h catches the vast majority of duplicates.
func NewMemoryStore(ttl time.Duration) *MemoryStore {
	s := &MemoryStore{
		entries: make(map[string]time.Time),
		ttl:     ttl,
		done:    make(chan struct{}),
	}
	go s.cleanup(ttl / 2)
	return s
}

// Exists checks if the event ID has been seen within the TTL window.
func (s *MemoryStore) Exists(_ context.Context, eventID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ts, ok := s.entries[eventID]
	if !ok {
		return false, nil
	}
	if time.Since(ts) > s.ttl {
		return false, nil
	}
	return true, nil
}

// Store records an event ID with the current timestamp.
func (s *MemoryStore) Store(_ context.Context, eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[eventID] = time.Now()
	return nil
}

// Close stops the background cleanup goroutine.
func (s *MemoryStore) Close() {
	close(s.done)
}

func (s *MemoryStore) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for id, ts := range s.entries {
				if now.Sub(ts) > s.ttl {
					delete(s.entries, id)
				}
			}
			s.mu.Unlock()
		case <-s.done:
			return
		}
	}
}
