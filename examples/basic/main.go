// Basic example: single webhook endpoint with sync processing.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	sw "github.com/hseinmoussa/shopify-webhook-go"
)

func main() {
	secret := os.Getenv("SHOPIFY_WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("SHOPIFY_WEBHOOK_SECRET is required")
	}

	router := sw.NewRouter(
		sw.WithErrorHandler(func(event sw.Event, err error) {
			log.Printf("ERROR [%s] shop=%s: %v",
				event.Metadata.Topic, event.Metadata.ShopDomain, err)
		}),
	)

	router.Handle(sw.TopicOrdersCreate, func(event sw.Event) error {
		var order sw.Order
		if err := event.Unmarshal(&order); err != nil {
			return fmt.Errorf("unmarshal order: %w", err)
		}
		log.Printf("New order #%d from %s — %s %s",
			order.OrderNumber, order.Email, order.TotalPrice, order.Currency)
		return nil
	})

	router.Handle(sw.TopicProductsUpdate, func(event sw.Event) error {
		var product sw.Product
		if err := event.Unmarshal(&product); err != nil {
			return fmt.Errorf("unmarshal product: %w", err)
		}
		log.Printf("Product updated: %s (%d variants)", product.Title, len(product.Variants))
		return nil
	})

	// GDPR mandatory webhooks.
	sw.RegisterGDPR(router, sw.GDPRHandlers{
		OnCustomerDataRequest: func(event sw.Event, p sw.CustomerDataRequest) error {
			log.Printf("GDPR data request for customer %d at %s", p.Customer.ID, p.ShopDomain)
			return nil
		},
		OnCustomerRedact: func(event sw.Event, p sw.CustomerRedact) error {
			log.Printf("GDPR redact customer %d at %s", p.Customer.ID, p.ShopDomain)
			return nil
		},
		OnShopRedact: func(event sw.Event, p sw.ShopRedact) error {
			log.Printf("GDPR shop redact for %s", p.ShopDomain)
			return nil
		},
	})

	handler := sw.Handler(secret, router)
	http.Handle("/webhooks", handler)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
