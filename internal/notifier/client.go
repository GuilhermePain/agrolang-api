package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
}

type notifyPayload struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

// Notify posts the alert message to the Evolution API webhook (RF-04.1).
func (c *Client) Notify(ctx context.Context, phone, message string) error {
	body, err := json.Marshal(notifyPayload{Phone: phone, Message: message})
	if err != nil {
		return fmt.Errorf("notifier: encode payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notifier: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notifier: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notifier: unexpected status %d", resp.StatusCode)
	}
	return nil
}
