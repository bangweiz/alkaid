package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/bangweiz/alkaid/internal/llm"
)

type Client struct {
}

func (c *Client) Chat(ctx context.Context, req llm.ChatRequest) (<-chan llm.ChatResponse, error) {
	stream := make(chan llm.ChatResponse)

	go func() {
		body, err := json.Marshal(req)
		if err != nil {
			close(stream)
		}

		client := &http.Client{
			Timeout: 1 * time.Minute,
		}

		res, err := client.Post("http://localhost:11434/api/chat", "application/json", bytes.NewReader(body))
		if err != nil {
			close(stream)
		}

		decoder := json.NewDecoder(res.Body)
		defer res.Body.Close()

		for {
			var chunk llm.ChatResponse

			err := decoder.Decode(&chunk)
			if errors.Is(err, io.EOF) {
				close(stream)
				break
			}
			if err != nil {
				close(stream)
				break
			}
			stream <- chunk
			if chunk.Done {
				close(stream)
				break
			}
		}
	}()

	return stream, nil
}
