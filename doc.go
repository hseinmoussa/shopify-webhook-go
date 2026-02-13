// Package shopifywebhook provides production-ready Shopify webhook handling for Go.
//
// It handles HMAC-SHA256 signature verification, topic-based routing,
// async processing (to meet Shopify's 5-second timeout), idempotency/dedup,
// and GDPR mandatory webhooks.
//
// Core features require zero external dependencies (stdlib only).
// Framework adapters for Gin, Echo, and Chi are available as separate modules.
//
// Basic usage:
//
//	secret := os.Getenv("SHOPIFY_WEBHOOK_SECRET")
//
//	router := shopifywebhook.NewRouter()
//	router.Handle(shopifywebhook.TopicOrdersCreate, func(event shopifywebhook.Event) error {
//	    var order shopifywebhook.Order
//	    if err := event.Unmarshal(&order); err != nil {
//	        return err
//	    }
//	    log.Printf("New order #%d", order.OrderNumber)
//	    return nil
//	})
//
//	handler := shopifywebhook.Handler(secret, router)
//	http.Handle("/webhooks", handler)
package shopifywebhook
