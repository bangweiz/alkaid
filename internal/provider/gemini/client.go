package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const (
	defaultAPIVersion  = "2026-05-20"
	defaultHTTPTimeout = 10 * time.Minute
)

type Client struct {
	baseURL     string
	apiKey      string
	apiRevision string
	httpClient  *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:     baseURL,
		apiKey:      apiKey,
		apiRevision: defaultAPIVersion,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
}

func (c *Client) CreateInteraction(ctx context.Context, req *InteractionRequest) (*InteractionResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", c.apiKey)
	httpReq.Header.Set("Api-Revision", c.apiRevision)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, err
	}

	var response InteractionResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
