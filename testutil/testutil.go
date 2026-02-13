// Package testutil provides helpers for testing Shopify webhook handlers.
//
// It generates properly signed HTTP requests that pass HMAC verification,
// so you can unit test your handlers without hitting Shopify.
package testutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"time"

	shopifywebhook "github.com/hseinmoussa/shopify-webhook-go"
)

var eventCounter atomic.Int64

// SignPayload returns the base64-encoded HMAC-SHA256 signature for the
// given secret and payload. Useful for custom test setups.
func SignPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// NewRequest creates a signed *http.Request suitable for testing webhook handlers.
//
//	req := testutil.NewRequest("test-secret", shopifywebhook.TopicOrdersCreate,
//	    "mystore.myshopify.com", []byte(`{"id":1}`))
//	rr := httptest.NewRecorder()
//	handler.ServeHTTP(rr, req)
func NewRequest(secret string, topic shopifywebhook.Topic, shopDomain string, body []byte) *http.Request {
	signature := SignPayload(secret, body)
	eventID := fmt.Sprintf("test-event-%d", eventCounter.Add(1))

	req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Topic", string(topic))
	req.Header.Set("X-Shopify-Hmac-Sha256", signature)
	req.Header.Set("X-Shopify-Shop-Domain", shopDomain)
	req.Header.Set("X-Shopify-Event-Id", eventID)
	req.Header.Set("X-Shopify-Webhook-Id", "test-webhook-id")
	req.Header.Set("X-Shopify-Triggered-At", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("X-Shopify-Api-Version", "2025-01")

	return req
}

// NewTypedRequest creates a signed request from a Go struct.
// The struct is marshaled to JSON, then signed.
//
//	order := shopifywebhook.Order{ID: 123, Email: "test@example.com"}
//	req := testutil.NewTypedRequest("secret", shopifywebhook.TopicOrdersCreate,
//	    "mystore.myshopify.com", order)
func NewTypedRequest(secret string, topic shopifywebhook.Topic, shopDomain string, payload any) *http.Request {
	body, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Sprintf("testutil: failed to marshal payload: %v", err))
	}
	return NewRequest(secret, topic, shopDomain, body)
}

// MustJSON marshals v to JSON bytes, panicking on error.
// Convenience for inline test payloads.
func MustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("testutil: failed to marshal: %v", err))
	}
	return b
}
