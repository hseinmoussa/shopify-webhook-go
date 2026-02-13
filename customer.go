package shopifywebhook

// Customer represents a Shopify customer webhook payload.
type Customer struct {
	ID                int64             `json:"id"`
	AdminGraphqlAPIID string            `json:"admin_graphql_api_id"`
	Email             string            `json:"email"`
	FirstName         string            `json:"first_name"`
	LastName          string            `json:"last_name"`
	Phone             string            `json:"phone"`
	State             string            `json:"state"`
	Note              string            `json:"note"`
	Tags              string            `json:"tags"`
	Currency          string            `json:"currency"`
	TaxExempt         bool              `json:"tax_exempt"`
	VerifiedEmail     bool              `json:"verified_email"`
	OrdersCount       int               `json:"orders_count"`
	TotalSpent        string            `json:"total_spent"`
	Addresses         []CustomerAddress `json:"addresses"`
	DefaultAddress    *CustomerAddress  `json:"default_address"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
}

// CustomerAddress represents a customer's address.
type CustomerAddress struct {
	ID           int64  `json:"id"`
	CustomerID   int64  `json:"customer_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Company      string `json:"company"`
	Address1     string `json:"address1"`
	Address2     string `json:"address2"`
	City         string `json:"city"`
	Province     string `json:"province"`
	ProvinceCode string `json:"province_code"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code"`
	Zip          string `json:"zip"`
	Phone        string `json:"phone"`
	Default      bool   `json:"default"`
}
