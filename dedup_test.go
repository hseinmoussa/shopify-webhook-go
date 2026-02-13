package shopifywebhook

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStore_StoreAndExists(t *testing.T) {
	store := NewMemoryStore(time.Hour)
	defer store.Close()
	ctx := context.Background()

	exists, err := store.Exists(ctx, "event-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatal("expected event-1 to not exist")
	}

	if err := store.Store(ctx, "event-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, err = store.Exists(ctx, "event-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatal("expected event-1 to exist after Store")
	}
}

func TestMemoryStore_TTLExpiry(t *testing.T) {
	store := NewMemoryStore(50 * time.Millisecond)
	defer store.Close()
	ctx := context.Background()

	_ = store.Store(ctx, "event-2")

	exists, _ := store.Exists(ctx, "event-2")
	if !exists {
		t.Fatal("expected event-2 to exist immediately after Store")
	}

	time.Sleep(100 * time.Millisecond)

	exists, _ = store.Exists(ctx, "event-2")
	if exists {
		t.Fatal("expected event-2 to have expired")
	}
}

func TestMemoryStore_Cleanup(t *testing.T) {
	// TTL=50ms, cleanup runs every 25ms.
	store := NewMemoryStore(50 * time.Millisecond)
	defer store.Close()
	ctx := context.Background()

	_ = store.Store(ctx, "event-3")
	_ = store.Store(ctx, "event-4")

	// Wait for TTL + cleanup interval.
	time.Sleep(150 * time.Millisecond)

	store.mu.RLock()
	count := len(store.entries)
	store.mu.RUnlock()

	if count != 0 {
		t.Fatalf("expected 0 entries after cleanup, got %d", count)
	}
}

func TestMemoryStore_MultipleEvents(t *testing.T) {
	store := NewMemoryStore(time.Hour)
	defer store.Close()
	ctx := context.Background()

	_ = store.Store(ctx, "a")
	_ = store.Store(ctx, "b")

	existsA, _ := store.Exists(ctx, "a")
	existsB, _ := store.Exists(ctx, "b")
	existsC, _ := store.Exists(ctx, "c")

	if !existsA || !existsB {
		t.Fatal("expected a and b to exist")
	}
	if existsC {
		t.Fatal("expected c to not exist")
	}
}
