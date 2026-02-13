package shopifywebhook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func signedRequest(secret, body string, topic Topic) *http.Request {
	sig := sign(secret, []byte(body))
	req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(body))
	req.Header.Set("X-Shopify-Hmac-Sha256", sig)
	req.Header.Set("X-Shopify-Topic", string(topic))
	req.Header.Set("X-Shopify-Shop-Domain", "test.myshopify.com")
	req.Header.Set("X-Shopify-Event-Id", "event-123")
	req.Header.Set("X-Shopify-Webhook-Id", "webhook-456")
	req.Header.Set("X-Shopify-Triggered-At", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("X-Shopify-Api-Version", "2025-01")
	return req
}

func TestMiddleware_Valid(t *testing.T) {
	secret := "test-secret"
	body := `{"id":1}`

	var gotEvent Event
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		event, ok := EventFromContext(r.Context())
		if !ok {
			t.Fatal("expected event in context")
		}
		gotEvent = event
		w.WriteHeader(http.StatusOK)
	})

	handler := Middleware(secret)(inner)
	req := signedRequest(secret, body, TopicOrdersCreate)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if gotEvent.Metadata.Topic != TopicOrdersCreate {
		t.Fatalf("expected topic %q, got %q", TopicOrdersCreate, gotEvent.Metadata.Topic)
	}
	if string(gotEvent.RawBody) != body {
		t.Fatalf("expected body %q, got %q", body, string(gotEvent.RawBody))
	}
}

func TestMiddleware_InvalidSignature(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := Middleware("correct-secret")(inner)

	req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(`{}`))
	req.Header.Set("X-Shopify-Hmac-Sha256", "aW52YWxpZA==")
	req.Header.Set("X-Shopify-Topic", "orders/create")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestMiddleware_MissingSignature(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := Middleware("secret")(inner)

	req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(`{}`))
	// No HMAC header.
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestMiddleware_BodyStillReadable(t *testing.T) {
	secret := "test-secret"
	body := `{"readable":true}`

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Body should still be readable after middleware.
		buf := make([]byte, 256)
		n, _ := r.Body.Read(buf)
		if string(buf[:n]) != body {
			t.Fatalf("expected body %q, got %q", body, string(buf[:n]))
		}
	})

	handler := Middleware(secret)(inner)
	req := signedRequest(secret, body, TopicOrdersCreate)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
}

func TestHandler_DispatchesSync(t *testing.T) {
	secret := "test-secret"
	body := `{"id":42}`

	var dispatched bool
	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error {
		dispatched = true
		return nil
	})

	handler := Handler(secret, router)
	req := signedRequest(secret, body, TopicOrdersCreate)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !dispatched {
		t.Fatal("expected handler to be dispatched")
	}
}

func TestHandler_InvalidSignature(t *testing.T) {
	router := NewRouter()
	handler := Handler("secret", router)

	req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(`{}`))
	req.Header.Set("X-Shopify-Hmac-Sha256", "bad")
	req.Header.Set("X-Shopify-Topic", "orders/create")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_WithDedup(t *testing.T) {
	secret := "test-secret"
	body := `{"id":1}`

	callCount := 0
	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error {
		callCount++
		return nil
	})

	store := NewMemoryStore(time.Hour)
	defer store.Close()

	handler := Handler(secret, router, WithIdempotencyStore(store))

	// First request — should be processed.
	req1 := signedRequest(secret, body, TopicOrdersCreate)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr1.Code)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 dispatch, got %d", callCount)
	}

	// Second request with same event ID — should be deduped.
	req2 := signedRequest(secret, body, TopicOrdersCreate)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
	if callCount != 1 {
		t.Fatalf("expected still 1 dispatch after dedup, got %d", callCount)
	}
}

func TestEventFromContext_Missing(t *testing.T) {
	_, ok := EventFromContext(context.Background())
	if ok {
		t.Fatal("expected ok=false for empty context")
	}
}
