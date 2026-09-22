package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

type StreamEventResult struct {
	Event any
	Error error
}

func (c *Client) CreateInteraction(ctx context.Context, req *InteractionRequest) (<-chan StreamEventResult, error) {
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

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini: api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	eventChan := make(chan StreamEventResult)
	go func() {
		defer resp.Body.Close()
		defer close(eventChan)

		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					eventChan <- StreamEventResult{Error: err}
				}
				return
			}

			line = strings.TrimSpace(line)

			if !strings.HasPrefix(line, "data:") {
				continue
			}

			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			if data == "[DONE]" {
				return
			}

			event, err := UnmarshalStreamEvent([]byte(data))
			if err != nil {
				eventChan <- StreamEventResult{Error: fmt.Errorf("gemini: parse event error: %w", err)}
				continue
			}

			eventChan <- StreamEventResult{Event: event}
		}
	}()

	return eventChan, nil
}
