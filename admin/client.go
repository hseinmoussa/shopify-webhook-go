// Package admin provides a client for managing Shopify webhook
// subscriptions via the REST Admin API.
package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Client manages webhook subscriptions via Shopify's REST Admin API.
type Client struct {
	shopDomain  string
	accessToken string
	apiVersion  string
	httpClient  *http.Client
}

// NewClient creates an Admin API client for the given shop.
//
//	client := admin.NewClient("mystore.myshopify.com", "shpat_xxx")
func NewClient(shopDomain, accessToken string, opts ...ClientOption) *Client {
	c := &Client{
		shopDomain:  shopDomain,
		accessToken: accessToken,
		apiVersion:  "2025-01",
		httpClient:  http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ClientOption configures the Admin API client.
type ClientOption func(*Client)

// WithAPIVersion sets the Shopify API version (e.g., "2025-01").
func WithAPIVersion(version string) ClientOption {
	return func(c *Client) { c.apiVersion = version }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = hc }
}

func (c *Client) baseURL() string {
	return fmt.Sprintf("https://%s/admin/api/%s", c.shopDomain, c.apiVersion)
}

// Create registers a new webhook subscription.
func (c *Client) Create(ctx context.Context, input WebhookInput) (*Webhook, error) {
	body, err := json.Marshal(webhookInputWrapper{Webhook: &input})
	if err != nil {
		return nil, fmt.Errorf("admin: marshal webhook input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL()+"/webhooks.json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("admin: create request: %w", err)
	}
	c.setHeaders(req)

	var result webhookWrapper
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Webhook, nil
}

// List returns webhook subscriptions, optionally filtered.
func (c *Client) List(ctx context.Context, opts *ListOptions) ([]Webhook, error) {
	u := c.baseURL() + "/webhooks.json"
	if opts != nil {
		u += "?" + opts.encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("admin: create request: %w", err)
	}
	c.setHeaders(req)

	var result webhooksWrapper
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Webhooks, nil
}

// Get retrieves a single webhook subscription by ID.
func (c *Client) Get(ctx context.Context, id int64) (*Webhook, error) {
	u := fmt.Sprintf("%s/webhooks/%d.json", c.baseURL(), id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("admin: create request: %w", err)
	}
	c.setHeaders(req)

	var result webhookWrapper
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Webhook, nil
}

// Update modifies an existing webhook subscription.
func (c *Client) Update(ctx context.Context, id int64, input WebhookInput) (*Webhook, error) {
	body, err := json.Marshal(webhookInputWrapper{Webhook: &input})
	if err != nil {
		return nil, fmt.Errorf("admin: marshal webhook input: %w", err)
	}

	u := fmt.Sprintf("%s/webhooks/%d.json", c.baseURL(), id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("admin: create request: %w", err)
	}
	c.setHeaders(req)

	var result webhookWrapper
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Webhook, nil
}

// Delete removes a webhook subscription.
func (c *Client) Delete(ctx context.Context, id int64) error {
	u := fmt.Sprintf("%s/webhooks/%d.json", c.baseURL(), id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return fmt.Errorf("admin: create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("admin: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.readError(resp)
	}
	return nil
}

// Count returns the number of webhook subscriptions.
func (c *Client) Count(ctx context.Context, opts *CountOptions) (int, error) {
	u := c.baseURL() + "/webhooks/count.json"
	if opts != nil {
		u += "?" + opts.encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, fmt.Errorf("admin: create request: %w", err)
	}
	c.setHeaders(req)

	var result countWrapper
	if err := c.do(req, &result); err != nil {
		return 0, err
	}
	return result.Count, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", c.accessToken)
}

func (c *Client) do(req *http.Request, v any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("admin: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.readError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("admin: decode response: %w", err)
	}
	return nil
}

func (c *Client) readError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return &APIError{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}
}

// APIError represents an error response from the Shopify Admin API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("admin: shopify API error (status %d): %s", e.StatusCode, e.Body)
}

func (o *ListOptions) encode() string {
	v := url.Values{}
	if o.Topic != "" {
		v.Set("topic", o.Topic)
	}
	if o.Address != "" {
		v.Set("address", o.Address)
	}
	if o.Limit > 0 {
		v.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.SinceID > 0 {
		v.Set("since_id", strconv.FormatInt(o.SinceID, 10))
	}
	return v.Encode()
}

func (o *CountOptions) encode() string {
	v := url.Values{}
	if o.Topic != "" {
		v.Set("topic", o.Topic)
	}
	if o.Address != "" {
		v.Set("address", o.Address)
	}
	return v.Encode()
}
