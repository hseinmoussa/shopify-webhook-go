package shopifywebhook

// CustomerDataRequest is the payload for customers/data_request webhooks.
// Shopify sends this when a customer requests their data under GDPR/CCPA.
type CustomerDataRequest struct {
	ShopID          int64             `json:"shop_id"`
	ShopDomain      string            `json:"shop_domain"`
	OrdersRequested []int64           `json:"orders_requested"`
	Customer        GDPRCustomer      `json:"customer"`
	DataRequest     GDPRDataRequestID `json:"data_request"`
}

// CustomerRedact is the payload for customers/redact webhooks.
// Shopify sends this when a customer requests deletion of their data.
type CustomerRedact struct {
	ShopID         int64        `json:"shop_id"`
	ShopDomain     string       `json:"shop_domain"`
	Customer       GDPRCustomer `json:"customer"`
	OrdersToRedact []int64      `json:"orders_to_redact"`
}

// ShopRedact is the payload for shop/redact webhooks.
// Shopify sends this 48 hours after a store owner uninstalls your app.
type ShopRedact struct {
	ShopID     int64  `json:"shop_id"`
	ShopDomain string `json:"shop_domain"`
}

// GDPRCustomer identifies the customer in GDPR webhook payloads.
type GDPRCustomer struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// GDPRDataRequestID identifies the data request.
type GDPRDataRequestID struct {
	ID int64 `json:"id"`
}

// GDPRHandlers groups the three mandatory GDPR webhook handlers.
// All three must be set. Using RegisterGDPR ensures none are omitted.
type GDPRHandlers struct {
	OnCustomerDataRequest func(event Event, payload CustomerDataRequest) error
	OnCustomerRedact      func(event Event, payload CustomerRedact) error
	OnShopRedact          func(event Event, payload ShopRedact) error
}

// RegisterGDPR registers all three mandatory GDPR handlers on the router.
// Panics if any handler in the GDPRHandlers struct is nil — this ensures
// developers don't accidentally omit a mandatory webhook.
func RegisterGDPR(router *Router, handlers GDPRHandlers) {
	if handlers.OnCustomerDataRequest == nil {
		panic("shopifywebhook: GDPRHandlers.OnCustomerDataRequest must not be nil")
	}
	if handlers.OnCustomerRedact == nil {
		panic("shopifywebhook: GDPRHandlers.OnCustomerRedact must not be nil")
	}
	if handlers.OnShopRedact == nil {
		panic("shopifywebhook: GDPRHandlers.OnShopRedact must not be nil")
	}

	router.Handle(TopicCustomersDataRequest, func(event Event) error {
		var payload CustomerDataRequest
		if err := event.Unmarshal(&payload); err != nil {
			return err
		}
		return handlers.OnCustomerDataRequest(event, payload)
	})

	router.Handle(TopicCustomersRedact, func(event Event) error {
		var payload CustomerRedact
		if err := event.Unmarshal(&payload); err != nil {
			return err
		}
		return handlers.OnCustomerRedact(event, payload)
	})

	router.Handle(TopicShopRedact, func(event Event) error {
		var payload ShopRedact
		if err := event.Unmarshal(&payload); err != nil {
			return err
		}
		return handlers.OnShopRedact(event, payload)
	})
}
