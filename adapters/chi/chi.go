// Package chi provides a Chi-compatible adapter for shopify-webhook-go.
//
// Since Chi uses standard net/http middleware, this is a thin wrapper
// around the core package. Provided for API consistency across adapters.
package chi

import (
	"net/http"

	shopifywebhook "github.com/hseinmoussa/shopify-webhook-go"
)

// Middleware returns Chi-compatible middleware for Shopify webhook verification.
// Delegates to the core package's Middleware since Chi uses standard net/http.
func Middleware(secret string, opts ...shopifywebhook.MiddlewareOption) func(http.Handler) http.Handler {
	return shopifywebhook.Middleware(secret, opts...)
}

// Handler returns a Chi-compatible http.Handler that verifies and dispatches webhooks.
func Handler(secret string, router *shopifywebhook.Router, opts ...shopifywebhook.HandlerOption) http.Handler {
	return shopifywebhook.Handler(secret, router, opts...)
}

// EventFromContext retrieves the Event from the request context.
func EventFromContext(r *http.Request) (shopifywebhook.Event, bool) {
	return shopifywebhook.EventFromContext(r.Context())
}
