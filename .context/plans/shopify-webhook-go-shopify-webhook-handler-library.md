# Plan: `shopify-webhook-go` — Shopify Webhook Handler Library for Go

## Context

Your GitHub profile needs showcase repos that demonstrate your backend/e-commerce expertise. There's no good Go library for production-grade Shopify webhook handling — `bold-commerce/go-shopify` only has basic HMAC verification, and `gnikyt/http_shopify_webhook` (7 stars) is too minimal. Developers re-implement verification, dedup, async processing, and GDPR compliance from scratch every time. This library fills that gap.

**Module:** `github.com/hseinmoussa/shopify-webhook-go`

---

## What the Library Provides

1. **HMAC-SHA256 verification** — reads raw body, verifies signature, returns bytes (solves the "parsed body breaks HMAC" bug)
2. **Topic-based router** — register handlers by webhook topic (`orders/create`, `products/update`, etc.)
3. **Type-safe payload structs** — Go structs for orders, products, customers, collections, carts, checkouts, refunds
4. **Async processing** — responds 200 immediately, processes in background worker pool (solves Shopify's 5-second timeout)
5. **Idempotency/dedup** — interface-based store tracking `X-Shopify-Event-Id`, with in-memory implementation included
6. **GDPR mandatory webhooks** — typed handlers for `customers/data_request`, `customers/redact`, `shop/redact` with enforcement that all three are registered
7. **Admin API client** — CRUD for webhook subscriptions via REST Admin API
8. **Framework adapters** — Gin, Echo, Chi (separate modules, no transitive deps)
9. **Test helpers** — generate signed test payloads for unit testing handlers

---

## File Structure

```
jackson/
├── go.mod
├── README.md (already exists — will become repo README)
├── LICENSE
├── doc.go
├── errors.go                  # Sentinel errors
├── webhook.go                 # Event, Metadata, Topic constants, HandlerFunc
├── verify.go                  # HMAC-SHA256 verification
├── verify_test.go
├── router.go                  # Topic-based handler routing
├── router_test.go
├── middleware.go              # net/http Middleware + all-in-one Handler
├── middleware_test.go
├── async.go                   # AsyncProcessor interface + WorkerPool
├── async_test.go
├── dedup.go                   # IdempotencyStore interface + MemoryStore
├── dedup_test.go
├── gdpr.go                    # GDPR types + RegisterGDPR
├── gdpr_test.go
├── order.go                   # Order, LineItem, ShippingLine, Fulfillment
├── product.go                 # Product, Variant, Image, ProductOption
├── customer.go                # Customer, CustomerAddress
├── collection.go              # Collection types
├── cart.go                    # Cart types
├── checkout.go                # Checkout types
├── refund.go                  # Refund types
├── common.go                  # Address, TaxLine, DiscountCode, shared types
├── admin/
│   ├── client.go              # REST Admin API client (CRUD webhooks)
│   └── webhook.go             # Webhook subscription types
├── adapters/
│   ├── gin/
│   │   ├── go.mod             # Separate module
│   │   └── gin.go
│   ├── echo/
│   │   ├── go.mod
│   │   └── echo.go
│   └── chi/
│       ├── go.mod
│       └── chi.go
├── testutil/
│   └── testutil.go            # SignPayload, NewRequest, NewTypedRequest
└── examples/
    ├── basic/main.go
    ├── async/main.go
    ├── gin-example/main.go
    └── admin-example/main.go
```

---

## Key Design Decisions

| Decision | Choice | Why |
|---|---|---|
| Core dependencies | **Zero** (stdlib only) | Adapter modules pull in Gin/Echo/Chi separately |
| Money fields | `string` not `decimal.Decimal` | Matches Shopify's JSON; no dependency needed |
| `Event.Unmarshal(&order)` | Method on Event struct | Type-safe at call site without generics |
| `Handle()` panics on duplicate topic | Yes | Catches config bugs at startup (Go idiom) |
| Worker pool when full | **Drop + call error handler** | Never block HTTP goroutine; Shopify retries |
| Dedup interface | `Exists` + `Store` (2 methods) | Minimal; easy to impl for Redis/Postgres |
| GDPR | Struct of 3 funcs, nil-check at registration | Forces all 3 mandatory handlers to be set |
| Handler default | Respond 200 immediately | Matches Shopify's 5s timeout requirement |

---

## Build Order

### Phase 1 — Foundation
1. `go.mod`, `doc.go`, `LICENSE`
2. `errors.go` — sentinel errors
3. `common.go` — Address, LineItem, TaxLine, shared types
4. `webhook.go` — Event, Metadata, Topic constants, ParseMetadata, HandlerFunc
5. `verify.go` + `verify_test.go` — HMAC verification
6. `router.go` + `router_test.go` — topic routing + dispatch

### Phase 2 — Middleware + Async
7. `middleware.go` + `middleware_test.go` — Middleware (composable) + Handler (all-in-one)
8. `async.go` + `async_test.go` — AsyncProcessor interface + WorkerPool
9. `dedup.go` + `dedup_test.go` — IdempotencyStore + MemoryStore
10. `gdpr.go` + `gdpr_test.go` — GDPR types + RegisterGDPR

### Phase 3 — Payload Structs
11. `order.go`, `product.go`, `customer.go`, `collection.go`, `cart.go`, `checkout.go`, `refund.go`

### Phase 4 — Admin Client + Test Helpers
12. `admin/client.go` + `admin/webhook.go`
13. `testutil/testutil.go`

### Phase 5 — Adapters
14. `adapters/gin/`, `adapters/echo/`, `adapters/chi/` (each with own `go.mod`)

### Phase 6 — Examples + README
15. `examples/` — working examples
16. `README.md` — quickstart, API docs, examples

---

## Verification

- `go build ./...` compiles all packages
- `go test ./...` passes all tests
- `go vet ./...` clean
- Verify test covers: valid HMAC, invalid HMAC, missing headers, topic routing, duplicate dedup, async dispatch, GDPR registration enforcement, worker pool shutdown
