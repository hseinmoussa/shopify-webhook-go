package shopifywebhook

// Refund represents a Shopify refund webhook payload.
type Refund struct {
	ID              int64            `json:"id"`
	OrderID         int64            `json:"order_id"`
	Note            string           `json:"note"`
	Restock         bool             `json:"restock"`
	UserID          int64            `json:"user_id"`
	RefundLineItems []RefundLineItem `json:"refund_line_items"`
	Transactions    []Transaction    `json:"transactions"`
	OrderAdjustments []OrderAdjustment `json:"order_adjustments"`
	CreatedAt       string           `json:"created_at"`
	ProcessedAt     string           `json:"processed_at"`
}

// RefundLineItem represents a line item being refunded.
type RefundLineItem struct {
	ID         int64    `json:"id"`
	LineItemID int64    `json:"line_item_id"`
	Quantity   int      `json:"quantity"`
	Subtotal   string   `json:"subtotal"`
	TotalTax   string   `json:"total_tax"`
	LineItem   LineItem `json:"line_item"`
}

// Transaction represents a payment transaction on a refund.
type Transaction struct {
	ID            int64  `json:"id"`
	OrderID       int64  `json:"order_id"`
	Kind          string `json:"kind"`
	Gateway       string `json:"gateway"`
	Status        string `json:"status"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Authorization string `json:"authorization"`
	ErrorCode     string `json:"error_code"`
	Message       string `json:"message"`
	CreatedAt     string `json:"created_at"`
}

// OrderAdjustment represents an adjustment on a refund (e.g., shipping refund).
type OrderAdjustment struct {
	ID           int64  `json:"id"`
	OrderID      int64  `json:"order_id"`
	RefundID     int64  `json:"refund_id"`
	Amount       string `json:"amount"`
	TaxAmount    string `json:"tax_amount"`
	Kind         string `json:"kind"`
	Reason       string `json:"reason"`
}
