package shopifywebhook

// Address represents a Shopify mailing address.
type Address struct {
	ID           int64   `json:"id,omitempty"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Company      string  `json:"company"`
	Address1     string  `json:"address1"`
	Address2     string  `json:"address2"`
	City         string  `json:"city"`
	Province     string  `json:"province"`
	ProvinceCode string  `json:"province_code"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
	Zip          string  `json:"zip"`
	Phone        string  `json:"phone"`
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
}

// LineItem represents an item in an order.
type LineItem struct {
	ID                  int64                `json:"id"`
	ProductID           int64                `json:"product_id"`
	VariantID           int64                `json:"variant_id"`
	Title               string               `json:"title"`
	VariantTitle        string               `json:"variant_title"`
	Quantity            int                  `json:"quantity"`
	Price               string               `json:"price"`
	SKU                 string               `json:"sku"`
	Vendor              string               `json:"vendor"`
	Grams               int64                `json:"grams"`
	Taxable             bool                 `json:"taxable"`
	RequiresShipping    bool                 `json:"requires_shipping"`
	GiftCard            bool                 `json:"gift_card"`
	FulfillmentStatus   string               `json:"fulfillment_status"`
	TaxLines            []TaxLine            `json:"tax_lines"`
	Properties          []NoteAttribute      `json:"properties"`
	DiscountAllocations []DiscountAllocation `json:"discount_allocations"`
}

// ShippingLine represents a shipping method applied to an order.
type ShippingLine struct {
	ID       int64     `json:"id"`
	Title    string    `json:"title"`
	Price    string    `json:"price"`
	Code     string    `json:"code"`
	Source   string    `json:"source"`
	TaxLines []TaxLine `json:"tax_lines"`
}

// TaxLine represents a tax applied to an order or line item.
type TaxLine struct {
	Title string  `json:"title"`
	Price string  `json:"price"`
	Rate  float64 `json:"rate"`
}

// DiscountCode represents a discount code applied to an order.
type DiscountCode struct {
	Code   string `json:"code"`
	Amount string `json:"amount"`
	Type   string `json:"type"`
}

// DiscountAllocation represents how a discount is allocated to a line item.
type DiscountAllocation struct {
	Amount                   string `json:"amount"`
	DiscountApplicationIndex int    `json:"discount_application_index"`
}

// NoteAttribute is a key-value pair attached to an order or line item.
type NoteAttribute struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// Fulfillment represents a fulfillment record for an order.
type Fulfillment struct {
	ID              int64      `json:"id"`
	OrderID         int64      `json:"order_id"`
	Status          string     `json:"status"`
	TrackingCompany string     `json:"tracking_company"`
	TrackingNumber  string     `json:"tracking_number"`
	TrackingNumbers []string   `json:"tracking_numbers"`
	TrackingURL     string     `json:"tracking_url"`
	TrackingURLs    []string   `json:"tracking_urls"`
	LineItems       []LineItem `json:"line_items"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}
