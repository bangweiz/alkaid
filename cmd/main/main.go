package main

import (
	"context"

	"github.com/bangweiz/alkaid/internal/llm"
	"github.com/bangweiz/alkaid/internal/llm/ollama"
	"github.com/bangweiz/alkaid/internal/tui"
)

func main() {
	var ollamaClient ollama.Client
	renderer, err := tui.NewRenderer()
	if err != nil {
		panic(err)
	}

	req := llm.ChatRequest{
		Model: "gemma4:26b",
		Messages: []llm.Message{
			{"user", "How to fire an HTTP call in Rust?"},
		},
	}

	stream, err := ollamaClient.Chat(context.Background(), req)
	if err != nil {
		panic(err)
	}

	renderer.RenderResponseStream(stream)
}
