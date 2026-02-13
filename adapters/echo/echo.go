// Package echo provides an Echo adapter for shopify-webhook-go.
package echo

import (
	"bytes"
	"io"

	"github.com/labstack/echo/v4"
	shopifywebhook "github.com/hseinmoussa/shopify-webhook-go"
)

const contextKey = "shopify_event"

// Middleware returns Echo middleware that verifies Shopify webhooks
// and stores the Event in the Echo context.
func Middleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			body, err := shopifywebhook.VerifyRequest(secret, c.Request())
			if err != nil {
				return c.NoContent(401)
			}

			meta, err := shopifywebhook.ParseMetadata(c.Request().Header)
			if err != nil {
				return c.NoContent(400)
			}

			event := shopifywebhook.Event{
				Metadata: meta,
				RawBody:  body,
			}

			c.Set(contextKey, event)
			c.Request().Body = io.NopCloser(bytes.NewReader(body))

			return next(c)
		}
	}
}

// EventFromContext retrieves the Event from Echo's context.
func EventFromContext(c echo.Context) (shopifywebhook.Event, bool) {
	val := c.Get(contextKey)
	if val == nil {
		return shopifywebhook.Event{}, false
	}
	event, ok := val.(shopifywebhook.Event)
	return event, ok
}
