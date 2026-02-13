package shopifywebhook

// Product represents a Shopify product webhook payload.
type Product struct {
	ID                int64           `json:"id"`
	AdminGraphqlAPIID string          `json:"admin_graphql_api_id"`
	Title             string          `json:"title"`
	Handle            string          `json:"handle"`
	BodyHTML          string          `json:"body_html"`
	Vendor            string          `json:"vendor"`
	ProductType       string          `json:"product_type"`
	Status            string          `json:"status"`
	Tags              string          `json:"tags"`
	TemplateSuffix    string          `json:"template_suffix"`
	Variants          []Variant       `json:"variants"`
	Images            []Image         `json:"images"`
	Options           []ProductOption `json:"options"`
	CreatedAt         string          `json:"created_at"`
	UpdatedAt         string          `json:"updated_at"`
	PublishedAt       string          `json:"published_at"`
}

// Variant represents a product variant.
type Variant struct {
	ID                  int64   `json:"id"`
	ProductID           int64   `json:"product_id"`
	Title               string  `json:"title"`
	Price               string  `json:"price"`
	CompareAtPrice      string  `json:"compare_at_price"`
	SKU                 string  `json:"sku"`
	Barcode             string  `json:"barcode"`
	Position            int     `json:"position"`
	Grams               int64   `json:"grams"`
	Weight              float64 `json:"weight"`
	WeightUnit          string  `json:"weight_unit"`
	InventoryItemID     int64   `json:"inventory_item_id"`
	InventoryQuantity   int     `json:"inventory_quantity"`
	InventoryManagement string  `json:"inventory_management"`
	InventoryPolicy     string  `json:"inventory_policy"`
	FulfillmentService  string  `json:"fulfillment_service"`
	Option1             string  `json:"option1"`
	Option2             string  `json:"option2"`
	Option3             string  `json:"option3"`
	Taxable             bool    `json:"taxable"`
	RequiresShipping    bool    `json:"requires_shipping"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

// Image represents a product image.
type Image struct {
	ID         int64   `json:"id"`
	ProductID  int64   `json:"product_id"`
	Position   int     `json:"position"`
	Src        string  `json:"src"`
	Alt        string  `json:"alt"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	VariantIDs []int64 `json:"variant_ids"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// ProductOption represents a product option (e.g., Size, Color).
type ProductOption struct {
	ID        int64    `json:"id"`
	ProductID int64    `json:"product_id"`
	Name      string   `json:"name"`
	Position  int      `json:"position"`
	Values    []string `json:"values"`
}
