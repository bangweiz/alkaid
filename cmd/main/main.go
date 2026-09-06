package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

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

	input := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := input.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintln(os.Stderr, "Read input:", err)
			os.Exit(1)
		}
		message := strings.TrimSpace(line)
		if message == "/exit" || (message == "" && errors.Is(err, io.EOF)) {
			return
		}
		if message == "" {
			continue
		}

		req := llm.ChatRequest{
			Model: "gemma4:26b",
			Messages: []llm.Message{
				{Role: "user", Content: message},
			},
		}

		stream, chatErr := ollamaClient.Chat(context.Background(), req)
		if chatErr != nil {
			panic(chatErr)
		}
		renderer.RenderResponseStream(stream)
		fmt.Println()
		if errors.Is(err, io.EOF) {
			return
		}
	}
}
