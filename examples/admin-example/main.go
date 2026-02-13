// Admin API example: register webhook subscriptions programmatically.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hseinmoussa/shopify-webhook-go/admin"
)

func main() {
	shopDomain := os.Getenv("SHOPIFY_SHOP_DOMAIN") // e.g., "mystore.myshopify.com"
	accessToken := os.Getenv("SHOPIFY_ACCESS_TOKEN")

	if shopDomain == "" || accessToken == "" {
		log.Fatal("SHOPIFY_SHOP_DOMAIN and SHOPIFY_ACCESS_TOKEN are required")
	}

	client := admin.NewClient(shopDomain, accessToken)
	ctx := context.Background()

	// Register a webhook for new orders.
	webhook, err := client.Create(ctx, admin.WebhookInput{
		Topic:   "orders/create",
		Address: "https://myapp.example.com/webhooks",
		Format:  "json",
	})
	if err != nil {
		log.Fatalf("Failed to create webhook: %v", err)
	}
	fmt.Printf("Created webhook %d for %s\n", webhook.ID, webhook.Topic)

	// List all registered webhooks.
	webhooks, err := client.List(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list webhooks: %v", err)
	}
	fmt.Printf("\n%d registered webhooks:\n", len(webhooks))
	for _, wh := range webhooks {
		fmt.Printf("  [%d] %s -> %s\n", wh.ID, wh.Topic, wh.Address)
	}

	// Count webhooks for a specific topic.
	count, err := client.Count(ctx, &admin.CountOptions{Topic: "orders/create"})
	if err != nil {
		log.Fatalf("Failed to count webhooks: %v", err)
	}
	fmt.Printf("\norders/create webhooks: %d\n", count)
}
