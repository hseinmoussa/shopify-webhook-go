# shopify-webhook-go

[![Go Reference](https://pkg.go.dev/badge/github.com/hseinmoussa/shopify-webhook-go.svg)](https://pkg.go.dev/github.com/hseinmoussa/shopify-webhook-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Production-ready Shopify webhook handling for Go. Verify signatures, route by topic, process async, deduplicate, and handle GDPR compliance — with zero external dependencies.

**Requires Go 1.21+**

## Why

Every Go developer building Shopify integrations re-implements the same things: HMAC verification, raw body parsing, deduplication, async processing to beat Shopify's 5-second timeout. This library handles all of it.

## Install

```bash
go get github.com/hseinmoussa/shopify-webhook-go
```

## Quick Start

```go
package main

import (
    "log"
    "net/http"
    "os"

    sw "github.com/hseinmoussa/shopify-webhook-go"
)

func main() {
    secret := os.Getenv("SHOPIFY_WEBHOOK_SECRET")

    router := sw.NewRouter()

    router.Handle(sw.TopicOrdersCreate, func(event sw.Event) error {
        var order sw.Order
        if err := event.Unmarshal(&order); err != nil {
            return err
        }
        log.Printf("New order #%d from %s — %s", order.OrderNumber, order.Email, order.TotalPrice)
        return nil
    })

    handler := sw.Handler(secret, router)
    http.Handle("/webhooks", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Features

### HMAC-SHA256 Verification

Reads the raw body before any parsing (solving the common "parsed body breaks HMAC" bug), verifies with constant-time comparison.

> **Note:** The `secret` is your Shopify app's **client secret** (found in app settings under "Client credentials"), not the access token.

```go
// As middleware (composable with any router)
mux.Handle("/webhooks", sw.Middleware(secret)(yourHandler))

// Or use the all-in-one Handler
mux.Handle("/webhooks", sw.Handler(secret, router))

// Or verify manually
body, err := sw.VerifyRequest(secret, r)
```

### Topic-Based Routing

Register handlers by Shopify webhook topic. Type constants for all standard topics.

```go
router := sw.NewRouter(
    sw.WithErrorHandler(func(event sw.Event, err error) {
        log.Printf("[%s] error: %v", event.Metadata.Topic, err)
    }),
)

router.Handle(sw.TopicOrdersCreate, handleNewOrder)
router.Handle(sw.TopicProductsUpdate, handleProductUpdate)
router.Handle(sw.TopicRefundsCreate, handleRefund)

// Catch-all for unregistered topics
router.Fallback(func(event sw.Event) error {
    log.Printf("unhandled topic: %s", event.Metadata.Topic)
    return nil
})
```

### Async Processing

Shopify drops webhooks that don't respond within 5 seconds. The `Handler` responds 200 immediately and processes in the background via a worker pool.

```go
pool := sw.NewWorkerPool(10, 1000,
    sw.WithPoolErrorHandler(func(event sw.Event, err error) {
        log.Printf("worker error: %v", err)
    }),
)
defer pool.Shutdown(context.Background())

handler := sw.Handler(secret, router,
    sw.WithAsyncProcessor(pool),
)
```

Implement `AsyncProcessor` to use your own queue (SQS, Kafka, Redis, etc.):

```go
type AsyncProcessor interface {
    Submit(event Event, router *Router)
    Shutdown(ctx context.Context) error
}
```

### Idempotency / Deduplication

Shopify can send the same webhook multiple times. Deduplicate using `X-Shopify-Event-Id`.

```go
store := sw.NewMemoryStore(24 * time.Hour) // In-memory, single instance
defer store.Close()

handler := sw.Handler(secret, router,
    sw.WithIdempotencyStore(store),
)
```

Implement `IdempotencyStore` for distributed deployments:

```go
type IdempotencyStore interface {
    Exists(ctx context.Context, eventID string) (bool, error)
    Store(ctx context.Context, eventID string) error
}
```

### GDPR Mandatory Webhooks

Shopify requires apps to handle three GDPR webhooks. `RegisterGDPR` enforces all three are set — panics at startup if any is nil.

```go
sw.RegisterGDPR(router, sw.GDPRHandlers{
    OnCustomerDataRequest: func(event sw.Event, p sw.CustomerDataRequest) error {
        // Handle data export request
        return nil
    },
    OnCustomerRedact: func(event sw.Event, p sw.CustomerRedact) error {
        // Delete customer data
        return nil
    },
    OnShopRedact: func(event sw.Event, p sw.ShopRedact) error {
        // Delete all shop data (48h after app uninstall)
        return nil
    },
})
```

### Type-Safe Payloads

Pre-built Go structs for common webhook topics. No more hand-rolling JSON tags.

```go
router.Handle(sw.TopicOrdersCreate, func(event sw.Event) error {
    var order sw.Order
    if err := event.Unmarshal(&order); err != nil {
        return err
    }
    // order.LineItems, order.Customer, order.ShippingAddress, etc.
    return nil
})
```

Available types: `Order`, `Product`, `Customer`, `Collection`, `Cart`, `Checkout`, `Refund` and all nested types (`LineItem`, `Variant`, `Address`, `Fulfillment`, etc.)

### Webhook Registration (Admin API)

Manage webhook subscriptions programmatically.

```go
import "github.com/hseinmoussa/shopify-webhook-go/admin"

client := admin.NewClient("mystore.myshopify.com", "shpat_xxx")

webhook, err := client.Create(ctx, admin.WebhookInput{
    Topic:   "orders/create",
    Address: "https://myapp.com/webhooks",
})
```

### Framework Adapters

Adapters for Gin, Echo, and Chi. Each is a separate module — importing the core library never pulls in framework dependencies.

```go
// Gin
import swgin "github.com/hseinmoussa/shopify-webhook-go/adapters/gin"
r.POST("/webhooks", swgin.Handler(secret, router))

// Echo
import swecho "github.com/hseinmoussa/shopify-webhook-go/adapters/echo"
e.POST("/webhooks", swecho.Handler(secret, router))

// Chi (uses standard net/http, thinnest wrapper)
import swchi "github.com/hseinmoussa/shopify-webhook-go/adapters/chi"
r.With(swchi.Middleware(secret)).Post("/webhooks", yourHandler)
```

### Test Helpers

Generate properly signed test requests for your webhook handlers.

```go
import "github.com/hseinmoussa/shopify-webhook-go/testutil"

func TestOrderWebhook(t *testing.T) {
    order := sw.Order{ID: 123, Email: "test@example.com", TotalPrice: "99.99"}
    req := testutil.NewTypedRequest("test-secret", sw.TopicOrdersCreate, "test.myshopify.com", order)

    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)

    assert.Equal(t, 200, rr.Code)
}
```

## Testing

### Unit tests

```bash
go test ./... -v
```

### Manual testing with curl

Start the example server:

```bash
SHOPIFY_WEBHOOK_SECRET=test-secret go run ./examples/basic
```

Send a signed webhook:

```bash
SECRET="test-secret"
BODY='{"id":1,"order_number":1001,"email":"customer@example.com","total_price":"99.99","currency":"USD"}'
SIG=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "$SECRET" -binary | base64)

curl -X POST http://localhost:8080/webhooks \
  -H "Content-Type: application/json" \
  -H "X-Shopify-Topic: orders/create" \
  -H "X-Shopify-Hmac-Sha256: $SIG" \
  -H "X-Shopify-Shop-Domain: test.myshopify.com" \
  -H "X-Shopify-Event-Id: test-event-1" \
  -H "X-Shopify-Webhook-Id: wh-1" \
  -H "X-Shopify-Triggered-At: 2025-01-01T00:00:00Z" \
  -H "X-Shopify-Api-Version: 2025-01" \
  -d "$BODY"
```

### Live testing with a Shopify store

1. Start the server with your app's **client secret**:
   ```bash
   SHOPIFY_WEBHOOK_SECRET=<your-client-secret> go run ./examples/basic
   ```

2. Expose it with [ngrok](https://ngrok.com):
   ```bash
   ngrok http 8080
   ```

3. Register a webhook on your store (replace placeholders):
   ```bash
   curl -X POST "https://YOUR-STORE.myshopify.com/admin/api/2025-01/webhooks.json" \
     -H "X-Shopify-Access-Token: YOUR_ACCESS_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"webhook":{"topic":"orders/create","address":"https://YOUR-NGROK-URL/webhooks","format":"json"}}'
   ```

4. Create a test order in your Shopify admin. You should see it logged.

5. **Clean up** when done — delete the webhook so Shopify doesn't keep hitting a dead URL:
   ```bash
   curl -X DELETE "https://YOUR-STORE.myshopify.com/admin/api/2025-01/webhooks/WEBHOOK_ID.json" \
     -H "X-Shopify-Access-Token: YOUR_ACCESS_TOKEN"
   ```

## Design Decisions

| Decision | Choice | Why |
|---|---|---|
| Dependencies | Zero (stdlib only) | Framework adapters are separate modules |
| Money fields | `string` | Matches Shopify's JSON; avoids decimal library dep |
| Async default | Respond 200 immediately | Shopify's 5-second timeout |
| Queue full | Drop + error callback | Never block HTTP; Shopify retries |
| Dedup interface | 2 methods (`Exists`/`Store`) | Easy to implement for Redis, Postgres, DynamoDB |
| GDPR | Panics on nil handler | Catches missing mandatory webhooks at startup |

## License

MIT
