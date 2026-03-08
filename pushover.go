package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const pushoverAPIBase = "https://api.pushover.net/1"

// PushoverClient handles communication with the Pushover API.
type PushoverClient struct {
	token   string
	userKey string
	http    *http.Client
	baseURL string
}

// NewPushoverClient creates a new PushoverClient.
func NewPushoverClient(token, userKey string, httpClient *http.Client) *PushoverClient {
	return &PushoverClient{
		token:   token,
		userKey: userKey,
		http:    httpClient,
		baseURL: pushoverAPIBase,
	}
}

// Configured returns an error if the client is missing required configuration.
func (c *PushoverClient) Configured() error {
	if c.token == "" {
		return fmt.Errorf("PUSHOVER_TOKEN environment variable is not set")
	}
	if c.userKey == "" {
		return fmt.Errorf("PUSHOVER_USER_KEY environment variable is not set")
	}
	return nil
}

// SendMessage sends a push notification via the Pushover API.
func (c *PushoverClient) SendMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	form := url.Values{}
	form.Set("token", c.token)
	form.Set("user", c.userKey)
	form.Set("message", req.Message)

	if req.Title != "" {
		form.Set("title", req.Title)
	}
	if req.Priority != 0 {
		form.Set("priority", strconv.Itoa(req.Priority))
	}
	if req.Sound != "" {
		form.Set("sound", req.Sound)
	}
	if req.Device != "" {
		form.Set("device", req.Device)
	}
	if req.URL != "" {
		form.Set("url", req.URL)
	}
	if req.URLTitle != "" {
		form.Set("url_title", req.URLTitle)
	}
	if req.HTML != 0 {
		form.Set("html", "1")
	}
	if req.Monospace != 0 {
		form.Set("monospace", "1")
	}
	if req.Timestamp != 0 {
		form.Set("timestamp", strconv.FormatInt(req.Timestamp, 10))
	}
	if req.TTL != 0 {
		form.Set("ttl", strconv.Itoa(req.TTL))
	}
	if req.Retry != 0 {
		form.Set("retry", strconv.Itoa(req.Retry))
	}
	if req.Expire != 0 {
		form.Set("expire", strconv.Itoa(req.Expire))
	}
	if req.Callback != "" {
		form.Set("callback", req.Callback)
	}
	if req.Tags != "" {
		form.Set("tags", req.Tags)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/messages.json", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending message: %w", err)
	}
	defer resp.Body.Close()

	var result MessageResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CheckReceipt polls the status of an emergency-priority receipt.
func (c *PushoverClient) CheckReceipt(ctx context.Context, receipt string) (*ReceiptResponse, error) {
	reqURL := fmt.Sprintf("%s/receipts/%s.json?token=%s", c.baseURL, receipt, c.token)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("checking receipt: %w", err)
	}
	defer resp.Body.Close()

	var result ReceiptResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CancelReceipt cancels an emergency-priority notification by receipt ID.
func (c *PushoverClient) CancelReceipt(ctx context.Context, receipt string) (*CancelResponse, error) {
	form := url.Values{}
	form.Set("token", c.token)

	endpoint := fmt.Sprintf("%s/receipts/%s/cancel.json", c.baseURL, receipt)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("canceling receipt: %w", err)
	}
	defer resp.Body.Close()

	var result CancelResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CancelReceiptByTag cancels all emergency-priority notifications matching a tag.
func (c *PushoverClient) CancelReceiptByTag(ctx context.Context, tag string) (*CancelResponse, error) {
	form := url.Values{}
	form.Set("token", c.token)

	endpoint := fmt.Sprintf("%s/receipts/cancel_by_tag/%s.json", c.baseURL, tag)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("canceling receipts by tag: %w", err)
	}
	defer resp.Body.Close()

	var result CancelResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func decodeResponse(resp *http.Response, v any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("decoding response (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
