package shopifywebhook

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

type contextKey int

const eventContextKey contextKey = iota

// EventFromContext retrieves the parsed Event from the request context.
// Returns the zero Event and false if not present.
func EventFromContext(ctx context.Context) (Event, bool) {
	event, ok := ctx.Value(eventContextKey).(Event)
	return event, ok
}

// Middleware returns net/http middleware that verifies Shopify webhook
// signatures and injects the parsed Event into the request context.
//
// It:
//  1. Reads the raw request body
//  2. Verifies the HMAC-SHA256 signature
//  3. Parses Shopify headers into Metadata
//  4. Stores the Event in the request context (retrieve with EventFromContext)
//  5. Replaces the request body so downstream handlers can still read it
//
// On verification failure, responds with 401 Unauthorized.
func Middleware(secret string, opts ...MiddlewareOption) func(http.Handler) http.Handler {
	cfg := &middlewareConfig{
		onVerifyError: func(w http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		},
		onParseError: func(w http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := VerifyRequest(secret, r)
			if err != nil {
				cfg.onVerifyError(w, r, err)
				return
			}

			meta, err := ParseMetadata(r.Header)
			if err != nil {
				cfg.onParseError(w, r, err)
				return
			}

			event := Event{
				Metadata: meta,
				RawBody:  body,
			}

			// Replace the body so downstream handlers can still read it.
			r.Body = io.NopCloser(bytes.NewReader(body))

			ctx := context.WithValue(r.Context(), eventContextKey, event)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Handler returns an http.Handler that verifies webhooks and dispatches
// events to the router. This is the all-in-one handler for a single
// webhook endpoint.
//
// It responds 200 OK immediately (to satisfy Shopify's 5-second timeout),
// then dispatches to the router synchronously or asynchronously depending
// on configuration.
func Handler(secret string, router *Router, opts ...HandlerOption) http.Handler {
	cfg := &handlerConfig{
		onVerifyError: func(w http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		},
		onParseError: func(w http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := VerifyRequest(secret, r)
		if err != nil {
			cfg.onVerifyError(w, r, err)
			return
		}

		meta, err := ParseMetadata(r.Header)
		if err != nil {
			cfg.onParseError(w, r, err)
			return
		}

		event := Event{
			Metadata: meta,
			RawBody:  body,
		}

		// Dedup check.
		if cfg.dedup != nil {
			processed, checkErr := cfg.dedup.Exists(r.Context(), event.Metadata.EventID)
			if checkErr == nil && processed {
				w.WriteHeader(http.StatusOK)
				return
			}
			// On dedup store errors, process anyway — better to duplicate
			// than to drop a webhook.
		}

		// Respond 200 immediately to satisfy Shopify's timeout.
		w.WriteHeader(http.StatusOK)

		if cfg.async != nil {
			cfg.async.Submit(event, router)
		} else {
			_ = router.Dispatch(event)
		}

		// Mark as processed after dispatch is submitted.
		if cfg.dedup != nil {
			_ = cfg.dedup.Store(context.Background(), event.Metadata.EventID)
		}
	})
}

// MiddlewareOption configures the verification Middleware.
type MiddlewareOption func(*middlewareConfig)

type middlewareConfig struct {
	onVerifyError func(http.ResponseWriter, *http.Request, error)
	onParseError  func(http.ResponseWriter, *http.Request, error)
}

// WithVerifyErrorHandler customizes the response when HMAC verification fails.
func WithVerifyErrorHandler(fn func(http.ResponseWriter, *http.Request, error)) MiddlewareOption {
	return func(c *middlewareConfig) {
		c.onVerifyError = fn
	}
}

// WithParseErrorHandler customizes the response when header parsing fails.
func WithParseErrorHandler(fn func(http.ResponseWriter, *http.Request, error)) MiddlewareOption {
	return func(c *middlewareConfig) {
		c.onParseError = fn
	}
}

// HandlerOption configures the all-in-one Handler.
type HandlerOption func(*handlerConfig)

type handlerConfig struct {
	async         AsyncProcessor
	dedup         IdempotencyStore
	onVerifyError func(http.ResponseWriter, *http.Request, error)
	onParseError  func(http.ResponseWriter, *http.Request, error)
}

// WithAsyncProcessor configures background event processing.
// When set, the Handler responds 200 immediately and dispatches
// the event to the processor in the background.
func WithAsyncProcessor(p AsyncProcessor) HandlerOption {
	return func(c *handlerConfig) {
		c.async = p
	}
}

// WithIdempotencyStore configures deduplication of webhook events
// using the X-Shopify-Event-Id header.
func WithIdempotencyStore(s IdempotencyStore) HandlerOption {
	return func(c *handlerConfig) {
		c.dedup = s
	}
}

// WithHandlerVerifyErrorHandler customizes the response when HMAC
// verification fails in the Handler.
func WithHandlerVerifyErrorHandler(fn func(http.ResponseWriter, *http.Request, error)) HandlerOption {
	return func(c *handlerConfig) {
		c.onVerifyError = fn
	}
}

// WithHandlerParseErrorHandler customizes the response when header
// parsing fails in the Handler.
func WithHandlerParseErrorHandler(fn func(http.ResponseWriter, *http.Request, error)) HandlerOption {
	return func(c *handlerConfig) {
		c.onParseError = fn
	}
}
