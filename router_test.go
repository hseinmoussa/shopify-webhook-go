package shopifywebhook

import (
	"errors"
	"fmt"
	"testing"
)

func TestRouter_Handle_And_Dispatch(t *testing.T) {
	router := NewRouter()

	var received Event
	router.Handle(TopicOrdersCreate, func(event Event) error {
		received = event
		return nil
	})

	event := Event{
		Metadata: Metadata{Topic: TopicOrdersCreate, ShopDomain: "test.myshopify.com"},
		RawBody:  []byte(`{"id":1}`),
	}

	if err := router.Dispatch(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Metadata.ShopDomain != "test.myshopify.com" {
		t.Fatalf("expected shop domain %q, got %q", "test.myshopify.com", received.Metadata.ShopDomain)
	}
}

func TestRouter_UnhandledTopic(t *testing.T) {
	router := NewRouter()

	event := Event{
		Metadata: Metadata{Topic: TopicOrdersCreate},
	}

	err := router.Dispatch(event)
	if !errors.Is(err, ErrUnhandledTopic) {
		t.Fatalf("expected ErrUnhandledTopic, got: %v", err)
	}
}

func TestRouter_Fallback(t *testing.T) {
	router := NewRouter()

	var fallbackTopic Topic
	router.Fallback(func(event Event) error {
		fallbackTopic = event.Metadata.Topic
		return nil
	})

	event := Event{
		Metadata: Metadata{Topic: "some/unknown_topic"},
	}

	if err := router.Dispatch(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fallbackTopic != "some/unknown_topic" {
		t.Fatalf("expected fallback to receive topic %q, got %q", "some/unknown_topic", fallbackTopic)
	}
}

func TestRouter_ErrorHandler(t *testing.T) {
	var capturedErr error
	router := NewRouter(
		WithErrorHandler(func(event Event, err error) {
			capturedErr = err
		}),
	)

	handlerErr := fmt.Errorf("processing failed")
	router.Handle(TopicProductsUpdate, func(event Event) error {
		return handlerErr
	})

	event := Event{
		Metadata: Metadata{Topic: TopicProductsUpdate},
	}

	err := router.Dispatch(event)
	if err == nil {
		t.Fatal("expected error from Dispatch")
	}
	if capturedErr != handlerErr {
		t.Fatalf("error handler got %v, want %v", capturedErr, handlerErr)
	}
}

func TestRouter_DuplicateHandle_Panics(t *testing.T) {
	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error { return nil })

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on duplicate Handle")
		}
	}()

	router.Handle(TopicOrdersCreate, func(event Event) error { return nil })
}

func TestRouter_Topics(t *testing.T) {
	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error { return nil })
	router.Handle(TopicProductsUpdate, func(event Event) error { return nil })

	topics := router.Topics()
	if len(topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(topics))
	}

	found := map[Topic]bool{}
	for _, tp := range topics {
		found[tp] = true
	}
	if !found[TopicOrdersCreate] || !found[TopicProductsUpdate] {
		t.Fatalf("expected both topics, got %v", topics)
	}
}

func TestRouter_HandlerReturnsNil(t *testing.T) {
	router := NewRouter()
	router.Handle(TopicOrdersCreate, func(event Event) error {
		return nil
	})

	event := Event{Metadata: Metadata{Topic: TopicOrdersCreate}}
	if err := router.Dispatch(event); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}
