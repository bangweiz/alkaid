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

	"github.com/bangweiz/alkaid/internal/llm"
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

// Stream implements the text-generation behavior needed by agents while
// keeping Gemini-specific request and event types inside this package.
func (c *Client) Stream(ctx context.Context, req llm.Request) (<-chan llm.StreamOutput, error) {
	geminiReq := toInteractionRequest(req)
	payload, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshal request: %w", err)
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
		return nil, fmt.Errorf("gemini: send request: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini: api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	eventChan := make(chan llm.StreamOutput)
	go func() {
		defer resp.Body.Close()
		defer close(eventChan)

		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					eventChan <- &StreamErrorEvent{cause: err}
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
				eventChan <- &StreamErrorEvent{cause: fmt.Errorf("gemini: parse event error: %w", err)}
				continue
			}

			if output, ok := event.(llm.StreamOutput); ok {
				eventChan <- output
			}
		}
	}()

	return eventChan, nil
}

func toInteractionRequest(req llm.Request) InteractionRequest {
	var systemInstruction string
	var userInput strings.Builder

	for _, message := range req.Messages {
		if message.Role == llm.RoleSystem {
			systemInstruction = message.Content
			continue
		}
		userInput.WriteString(message.Content)
		userInput.WriteByte('\n')
	}

	return InteractionRequest{
		Model:                 Model(req.Model),
		SystemInstruction:     systemInstruction,
		Input:                 strings.TrimSpace(userInput.String()),
		PreviousInteractionID: req.SessionID,
		Stream:                true,
	}
}
