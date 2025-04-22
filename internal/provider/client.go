package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	params     url.Values
}

type Option func(c *Client)

func NewClient(opts ...Option) *Client {
	client := &Client{
		httpClient: http.DefaultClient,
		params:     make(url.Values),
	}

	for _, o := range opts {
		o(client)
	}

	return client
}

func WithBaseURL(rawURL string) Option {
	return func(c *Client) {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			// In a real application, you might want to handle this error differently
			panic(err)
		}
		c.baseURL = parsed
	}
}

func WithQueryParam(key, value string) Option {
	return func(c *Client) {
		c.params.Add(key, value)
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func Fetch[T any](ctx context.Context, client *Client) (T, error) {
	var zero T

	// Create a copy of the base URL
	reqURL := *client.baseURL

	// Add query parameters
	reqURL.RawQuery = client.params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return zero, err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, err
	}

	return result, nil
}
