// Package gin provides a Gin adapter for shopify-webhook-go.
package gin

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	shopifywebhook "github.com/hseinmoussa/shopify-webhook-go"
)

const contextKey = "shopify_event"

// Middleware returns Gin middleware that verifies Shopify webhooks
// and stores the Event in the Gin context.
func Middleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := shopifywebhook.VerifyRequest(secret, c.Request)
		if err != nil {
			c.AbortWithStatus(401)
			return
		}

		meta, err := shopifywebhook.ParseMetadata(c.Request.Header)
		if err != nil {
			c.AbortWithStatus(400)
			return
		}

		event := shopifywebhook.Event{
			Metadata: meta,
			RawBody:  body,
		}

		c.Set(contextKey, event)
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Next()
	}
}

// EventFromContext retrieves the Event from Gin's context.
func EventFromContext(c *gin.Context) (shopifywebhook.Event, bool) {
	val, exists := c.Get(contextKey)
	if !exists {
		return shopifywebhook.Event{}, false
	}
	event, ok := val.(shopifywebhook.Event)
	return event, ok
}

// Handler returns a Gin handler that verifies and dispatches webhooks
// to the router. Wraps the core package's Handler.
func Handler(secret string, router *shopifywebhook.Router, opts ...shopifywebhook.HandlerOption) gin.HandlerFunc {
	h := shopifywebhook.Handler(secret, router, opts...)
	return gin.WrapH(h)
}
