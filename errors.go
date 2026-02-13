package shopifywebhook

import "errors"

var (
	// ErrInvalidSignature is returned when HMAC-SHA256 verification fails.
	ErrInvalidSignature = errors.New("shopifywebhook: invalid HMAC signature")

	// ErrMissingSignature is returned when the X-Shopify-Hmac-Sha256 header is absent.
	ErrMissingSignature = errors.New("shopifywebhook: missing X-Shopify-Hmac-Sha256 header")

	// ErrMissingTopic is returned when the X-Shopify-Topic header is absent.
	ErrMissingTopic = errors.New("shopifywebhook: missing X-Shopify-Topic header")

	// ErrUnhandledTopic is returned when no handler is registered for a topic
	// and no fallback handler is set.
	ErrUnhandledTopic = errors.New("shopifywebhook: unhandled topic")

	// ErrQueueFull is returned when the async worker pool's queue is full
	// and the event is dropped.
	ErrQueueFull = errors.New("shopifywebhook: worker pool queue full, event dropped")
)
