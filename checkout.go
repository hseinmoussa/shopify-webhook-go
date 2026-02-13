package shopifywebhook

// Checkout represents a Shopify checkout webhook payload.
type Checkout struct {
	ID                    int64          `json:"id"`
	Token                 string         `json:"token"`
	CartToken             string         `json:"cart_token"`
	Email                 string         `json:"email"`
	Gateway               string         `json:"gateway"`
	TotalPrice            string         `json:"total_price"`
	SubtotalPrice         string         `json:"subtotal_price"`
	TotalTax              string         `json:"total_tax"`
	TotalDiscounts        string         `json:"total_discounts"`
	TotalWeight           int64          `json:"total_weight"`
	Currency              string         `json:"currency"`
	CompletedAt           string         `json:"completed_at"`
	Phone                 string         `json:"phone"`
	CustomerLocale        string         `json:"customer_locale"`
	LandingSite           string         `json:"landing_site"`
	ReferringSite         string         `json:"referring_site"`
	SourceName            string         `json:"source_name"`
	BuyerAcceptsMarketing bool           `json:"buyer_accepts_marketing"`
	TaxesIncluded         bool           `json:"taxes_included"`
	Customer              *Customer      `json:"customer"`
	LineItems             []LineItem     `json:"line_items"`
	ShippingLine          *ShippingLine  `json:"shipping_line"`
	BillingAddress        *Address       `json:"billing_address"`
	ShippingAddress       *Address       `json:"shipping_address"`
	DiscountCodes         []DiscountCode `json:"discount_codes"`
	TaxLines              []TaxLine      `json:"tax_lines"`
	NoteAttributes        []NoteAttribute `json:"note_attributes"`
	Note                  string         `json:"note"`
	CreatedAt             string         `json:"created_at"`
	UpdatedAt             string         `json:"updated_at"`
}
