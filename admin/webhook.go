package admin

// Webhook represents a Shopify webhook subscription resource.
type Webhook struct {
	ID                  int64    `json:"id"`
	Address             string   `json:"address"`
	Topic               string   `json:"topic"`
	Format              string   `json:"format"`
	Fields              []string `json:"fields,omitempty"`
	MetafieldNamespaces []string `json:"metafield_namespaces,omitempty"`
	APIVersion          string   `json:"api_version"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

// WebhookInput is used when creating or updating webhook subscriptions.
type WebhookInput struct {
	Address             string   `json:"address"`
	Topic               string   `json:"topic"`
	Format              string   `json:"format,omitempty"`
	Fields              []string `json:"fields,omitempty"`
	MetafieldNamespaces []string `json:"metafield_namespaces,omitempty"`
}

// ListOptions filters the List call.
type ListOptions struct {
	Topic   string
	Address string
	Limit   int
	SinceID int64
}

// CountOptions filters the Count call.
type CountOptions struct {
	Topic   string
	Address string
}

// webhookWrapper wraps a single webhook for JSON encoding/decoding.
type webhookWrapper struct {
	Webhook *Webhook `json:"webhook"`
}

// webhookInputWrapper wraps input for create/update requests.
type webhookInputWrapper struct {
	Webhook *WebhookInput `json:"webhook"`
}

// webhooksWrapper wraps a list of webhooks.
type webhooksWrapper struct {
	Webhooks []Webhook `json:"webhooks"`
}

// countWrapper wraps the count response.
type countWrapper struct {
	Count int `json:"count"`
}
