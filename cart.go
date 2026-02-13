package shopifywebhook

// Cart represents a Shopify cart webhook payload.
type Cart struct {
	ID        string         `json:"id"`
	Token     string         `json:"token"`
	Note      string         `json:"note"`
	LineItems []CartLineItem `json:"line_items"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

// CartLineItem represents an item in a cart.
type CartLineItem struct {
	ID         int64           `json:"id"`
	ProductID  int64           `json:"product_id"`
	VariantID  int64           `json:"variant_id"`
	Title      string          `json:"title"`
	Quantity   int             `json:"quantity"`
	Price      string          `json:"price"`
	SKU        string          `json:"sku"`
	Grams      int64           `json:"grams"`
	Vendor     string          `json:"vendor"`
	Properties []NoteAttribute `json:"properties"`
}
