package shopifywebhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Topic represents a Shopify webhook topic.
type Topic string

// Webhook topics for orders.
const (
	TopicOrdersCreate             Topic = "orders/create"
	TopicOrdersUpdate             Topic = "orders/update"
	TopicOrdersDelete             Topic = "orders/delete"
	TopicOrdersCancelled          Topic = "orders/cancelled"
	TopicOrdersFulfilled          Topic = "orders/fulfilled"
	TopicOrdersPaid               Topic = "orders/paid"
	TopicOrdersPartiallyFulfilled Topic = "orders/partially_fulfilled"
)

// Webhook topics for products.
const (
	TopicProductsCreate Topic = "products/create"
	TopicProductsUpdate Topic = "products/update"
	TopicProductsDelete Topic = "products/delete"
)

// Webhook topics for customers.
const (
	TopicCustomersCreate  Topic = "customers/create"
	TopicCustomersUpdate  Topic = "customers/update"
	TopicCustomersDelete  Topic = "customers/delete"
	TopicCustomersEnable  Topic = "customers/enable"
	TopicCustomersDisable Topic = "customers/disable"
)

// Webhook topics for collections.
const (
	TopicCollectionsCreate Topic = "collections/create"
	TopicCollectionsUpdate Topic = "collections/update"
	TopicCollectionsDelete Topic = "collections/delete"
)

// Webhook topics for carts.
const (
	TopicCartsCreate Topic = "carts/create"
	TopicCartsUpdate Topic = "carts/update"
)

// Webhook topics for checkouts.
const (
	TopicCheckoutsCreate Topic = "checkouts/create"
	TopicCheckoutsUpdate Topic = "checkouts/update"
	TopicCheckoutsDelete Topic = "checkouts/delete"
)

// Webhook topics for refunds.
const (
	TopicRefundsCreate Topic = "refunds/create"
)

// Webhook topics for app lifecycle.
const (
	TopicAppUninstalled Topic = "app/uninstalled"
)

// GDPR mandatory webhook topics.
const (
	TopicCustomersDataRequest Topic = "customers/data_request"
	TopicCustomersRedact      Topic = "customers/redact"
	TopicShopRedact           Topic = "shop/redact"
)

// Metadata contains the Shopify headers extracted from a webhook request.
type Metadata struct {
	Topic       Topic
	HmacSHA256  string
	ShopDomain  string
	WebhookID   string
	EventID     string
	TriggeredAt time.Time
	APIVersion  string
}

// Event represents a parsed and verified Shopify webhook event.
type Event struct {
	Metadata Metadata
	RawBody  []byte
}

// Unmarshal decodes the raw body into the provided Go value.
//
//	var order shopifywebhook.Order
//	if err := event.Unmarshal(&order); err != nil { ... }
func (e *Event) Unmarshal(v any) error {
	return json.Unmarshal(e.RawBody, v)
}

// HandlerFunc is the function signature for webhook topic handlers.
type HandlerFunc func(event Event) error

// ErrorHandlerFunc is called when a HandlerFunc returns an error.
type ErrorHandlerFunc func(event Event, err error)

// ParseMetadata extracts Shopify webhook metadata from HTTP request headers.
// Returns an error if required headers (Topic, HmacSHA256) are missing.
func ParseMetadata(h http.Header) (Metadata, error) {
	topic := h.Get("X-Shopify-Topic")
	if topic == "" {
		return Metadata{}, ErrMissingTopic
	}

	hmacHeader := h.Get("X-Shopify-Hmac-Sha256")
	if hmacHeader == "" {
		return Metadata{}, ErrMissingSignature
	}

	var triggeredAt time.Time
	if ts := h.Get("X-Shopify-Triggered-At"); ts != "" {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			triggeredAt = parsed
		}
	}

	return Metadata{
		Topic:       Topic(topic),
		HmacSHA256:  hmacHeader,
		ShopDomain:  h.Get("X-Shopify-Shop-Domain"),
		WebhookID:   h.Get("X-Shopify-Webhook-Id"),
		EventID:     h.Get("X-Shopify-Event-Id"),
		TriggeredAt: triggeredAt,
		APIVersion:  h.Get("X-Shopify-Api-Version"),
	}, nil
}

// String returns the topic string.
func (t Topic) String() string {
	return string(t)
}

// Validate checks if the topic matches a known Shopify webhook topic.
// Returns an error with the topic name if unknown.
// Note: unknown topics are still routable — this is for advisory use only.
func (t Topic) Validate() error {
	switch t {
	case TopicOrdersCreate, TopicOrdersUpdate, TopicOrdersDelete,
		TopicOrdersCancelled, TopicOrdersFulfilled, TopicOrdersPaid,
		TopicOrdersPartiallyFulfilled,
		TopicProductsCreate, TopicProductsUpdate, TopicProductsDelete,
		TopicCustomersCreate, TopicCustomersUpdate, TopicCustomersDelete,
		TopicCustomersEnable, TopicCustomersDisable,
		TopicCollectionsCreate, TopicCollectionsUpdate, TopicCollectionsDelete,
		TopicCartsCreate, TopicCartsUpdate,
		TopicCheckoutsCreate, TopicCheckoutsUpdate, TopicCheckoutsDelete,
		TopicRefundsCreate,
		TopicAppUninstalled,
		TopicCustomersDataRequest, TopicCustomersRedact, TopicShopRedact:
		return nil
	default:
		return fmt.Errorf("shopifywebhook: unknown topic %q", t)
	}
}
