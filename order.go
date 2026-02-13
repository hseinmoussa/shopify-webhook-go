package shopifywebhook

// Order represents a Shopify order webhook payload.
type Order struct {
	ID                    int64            `json:"id"`
	AdminGraphqlAPIID     string           `json:"admin_graphql_api_id"`
	Email                 string           `json:"email"`
	Name                  string           `json:"name"`
	Number                int              `json:"number"`
	OrderNumber           int              `json:"order_number"`
	Note                  string           `json:"note"`
	Token                 string           `json:"token"`
	Gateway               string           `json:"gateway"`
	TotalPrice            string           `json:"total_price"`
	SubtotalPrice         string           `json:"subtotal_price"`
	TotalTax              string           `json:"total_tax"`
	TotalDiscounts        string           `json:"total_discounts"`
	TotalWeight           int64            `json:"total_weight"`
	Currency              string           `json:"currency"`
	FinancialStatus       string           `json:"financial_status"`
	FulfillmentStatus     string           `json:"fulfillment_status"`
	Confirmed             bool             `json:"confirmed"`
	Test                  bool             `json:"test"`
	CancelReason          string           `json:"cancel_reason"`
	Tags                  string           `json:"tags"`
	ContactEmail          string           `json:"contact_email"`
	Phone                 string           `json:"phone"`
	BrowserIP             string           `json:"browser_ip"`
	LandingSite           string           `json:"landing_site"`
	ReferringSite         string           `json:"referring_site"`
	SourceName            string           `json:"source_name"`
	Customer              *Customer        `json:"customer"`
	LineItems             []LineItem       `json:"line_items"`
	ShippingLines         []ShippingLine   `json:"shipping_lines"`
	BillingAddress        *Address         `json:"billing_address"`
	ShippingAddress       *Address         `json:"shipping_address"`
	Fulfillments          []Fulfillment    `json:"fulfillments"`
	Refunds               []Refund         `json:"refunds"`
	DiscountCodes         []DiscountCode   `json:"discount_codes"`
	NoteAttributes        []NoteAttribute  `json:"note_attributes"`
	TaxLines              []TaxLine        `json:"tax_lines"`
	PaymentGatewayNames   []string         `json:"payment_gateway_names"`
	CreatedAt             string           `json:"created_at"`
	UpdatedAt             string           `json:"updated_at"`
	ClosedAt              string           `json:"closed_at"`
	CancelledAt           string           `json:"cancelled_at"`
	ProcessedAt           string           `json:"processed_at"`
}
