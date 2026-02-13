package shopifywebhook

// Collection represents a Shopify collection webhook payload.
// Covers both custom collections and smart collections.
type Collection struct {
	ID                int64  `json:"id"`
	AdminGraphqlAPIID string `json:"admin_graphql_api_id"`
	Title             string `json:"title"`
	Handle            string `json:"handle"`
	BodyHTML          string `json:"body_html"`
	SortOrder         string `json:"sort_order"`
	TemplateSuffix    string `json:"template_suffix"`
	PublishedScope    string `json:"published_scope"`
	UpdatedAt         string `json:"updated_at"`
	PublishedAt       string `json:"published_at"`
}
