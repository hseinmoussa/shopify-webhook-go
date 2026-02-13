package shopifywebhook

import (
	"encoding/json"
	"testing"
)

func TestRegisterGDPR_AllHandlersCalled(t *testing.T) {
	router := NewRouter()

	var (
		dataRequestCalled bool
		customerRedacted  bool
		shopRedacted      bool
	)

	RegisterGDPR(router, GDPRHandlers{
		OnCustomerDataRequest: func(event Event, p CustomerDataRequest) error {
			dataRequestCalled = true
			if p.Customer.Email != "john@example.com" {
				t.Fatalf("expected email %q, got %q", "john@example.com", p.Customer.Email)
			}
			return nil
		},
		OnCustomerRedact: func(event Event, p CustomerRedact) error {
			customerRedacted = true
			if p.Customer.ID != 456 {
				t.Fatalf("expected customer ID 456, got %d", p.Customer.ID)
			}
			return nil
		},
		OnShopRedact: func(event Event, p ShopRedact) error {
			shopRedacted = true
			if p.ShopDomain != "closing-store.myshopify.com" {
				t.Fatalf("expected shop domain %q, got %q", "closing-store.myshopify.com", p.ShopDomain)
			}
			return nil
		},
	})

	// Test customers/data_request.
	dataReqPayload, _ := json.Marshal(CustomerDataRequest{
		ShopID:     1,
		ShopDomain: "test.myshopify.com",
		Customer:   GDPRCustomer{ID: 123, Email: "john@example.com"},
	})
	_ = router.Dispatch(Event{
		Metadata: Metadata{Topic: TopicCustomersDataRequest},
		RawBody:  dataReqPayload,
	})

	// Test customers/redact.
	redactPayload, _ := json.Marshal(CustomerRedact{
		ShopID:   1,
		Customer: GDPRCustomer{ID: 456},
	})
	_ = router.Dispatch(Event{
		Metadata: Metadata{Topic: TopicCustomersRedact},
		RawBody:  redactPayload,
	})

	// Test shop/redact.
	shopPayload, _ := json.Marshal(ShopRedact{
		ShopID:     1,
		ShopDomain: "closing-store.myshopify.com",
	})
	_ = router.Dispatch(Event{
		Metadata: Metadata{Topic: TopicShopRedact},
		RawBody:  shopPayload,
	})

	if !dataRequestCalled {
		t.Fatal("OnCustomerDataRequest was not called")
	}
	if !customerRedacted {
		t.Fatal("OnCustomerRedact was not called")
	}
	if !shopRedacted {
		t.Fatal("OnShopRedact was not called")
	}
}

func TestRegisterGDPR_NilHandlerPanics(t *testing.T) {
	tests := []struct {
		name     string
		handlers GDPRHandlers
	}{
		{
			name: "nil OnCustomerDataRequest",
			handlers: GDPRHandlers{
				OnCustomerDataRequest: nil,
				OnCustomerRedact:      func(Event, CustomerRedact) error { return nil },
				OnShopRedact:          func(Event, ShopRedact) error { return nil },
			},
		},
		{
			name: "nil OnCustomerRedact",
			handlers: GDPRHandlers{
				OnCustomerDataRequest: func(Event, CustomerDataRequest) error { return nil },
				OnCustomerRedact:      nil,
				OnShopRedact:          func(Event, ShopRedact) error { return nil },
			},
		},
		{
			name: "nil OnShopRedact",
			handlers: GDPRHandlers{
				OnCustomerDataRequest: func(Event, CustomerDataRequest) error { return nil },
				OnCustomerRedact:      func(Event, CustomerRedact) error { return nil },
				OnShopRedact:          nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("expected panic for nil handler")
				}
			}()
			RegisterGDPR(NewRouter(), tt.handlers)
		})
	}
}

func TestRegisterGDPR_UnmarshalError(t *testing.T) {
	router := NewRouter()
	RegisterGDPR(router, GDPRHandlers{
		OnCustomerDataRequest: func(Event, CustomerDataRequest) error { return nil },
		OnCustomerRedact:      func(Event, CustomerRedact) error { return nil },
		OnShopRedact:          func(Event, ShopRedact) error { return nil },
	})

	err := router.Dispatch(Event{
		Metadata: Metadata{Topic: TopicCustomersDataRequest},
		RawBody:  []byte(`not valid json`),
	})
	if err == nil {
		t.Fatal("expected unmarshal error for invalid JSON")
	}
}
