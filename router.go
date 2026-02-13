package shopifywebhook

import (
	"fmt"
	"sync"
)

// Router dispatches webhook events to registered handlers by topic.
type Router struct {
	mu       sync.RWMutex
	handlers map[Topic]HandlerFunc
	fallback HandlerFunc
	onError  ErrorHandlerFunc
}

// NewRouter creates a new Router with the given options.
func NewRouter(opts ...RouterOption) *Router {
	r := &Router{
		handlers: make(map[Topic]HandlerFunc),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Handle registers a handler for a specific webhook topic.
// Panics if a handler is already registered for the topic — this catches
// configuration mistakes at startup.
func (r *Router) Handle(topic Topic, handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.handlers[topic]; exists {
		panic(fmt.Sprintf("shopifywebhook: handler already registered for topic %q", topic))
	}
	r.handlers[topic] = handler
}

// Fallback sets a handler for topics without a registered handler.
// If not set, unhandled topics cause Dispatch to return ErrUnhandledTopic.
func (r *Router) Fallback(handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallback = handler
}

// Dispatch routes an event to the appropriate handler based on its topic.
//
// This is called internally by the Handler, but is also exported for use
// outside HTTP contexts (e.g., replaying events from a database or queue).
func (r *Router) Dispatch(event Event) error {
	r.mu.RLock()
	handler, ok := r.handlers[event.Metadata.Topic]
	fallback := r.fallback
	onError := r.onError
	r.mu.RUnlock()

	if !ok {
		if fallback != nil {
			handler = fallback
		} else {
			return fmt.Errorf("%w: %s", ErrUnhandledTopic, event.Metadata.Topic)
		}
	}

	if err := handler(event); err != nil {
		if onError != nil {
			onError(event, err)
		}
		return err
	}
	return nil
}

// Topics returns a list of all registered topics.
func (r *Router) Topics() []Topic {
	r.mu.RLock()
	defer r.mu.RUnlock()
	topics := make([]Topic, 0, len(r.handlers))
	for t := range r.handlers {
		topics = append(topics, t)
	}
	return topics
}

// RouterOption configures a Router.
type RouterOption func(*Router)

// WithErrorHandler sets the function called when a handler returns an error.
func WithErrorHandler(fn ErrorHandlerFunc) RouterOption {
	return func(r *Router) {
		r.onError = fn
	}
}
