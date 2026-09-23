package agents

import (
	"context"
	"fmt"

	"github.com/bangweiz/alkaid/internal/llm"
	"github.com/bangweiz/alkaid/internal/llm/gemini"
)

type CodingAgent struct {
	client llm.Client
}

func NewCodingAgent(client llm.Client) *CodingAgent {
	return &CodingAgent{client: client}
}

func (a *CodingAgent) Prompt(ctx context.Context, text string) error {
	req := llm.Request{
		Model: string(gemini.ModelGemini35FlashLite),
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "You are a Go specialist."},
			{Role: llm.RoleUser, Content: text},
		},
	}

	stream, err := a.client.Stream(ctx, req)
	if err != nil {
		return err
	}

	for chunk := range stream {
		if err := chunk.Err(); err != nil {
			return err
		}
		if chunk.Done() {
			_, outputTokens := chunk.TokenUsage()
			fmt.Printf("\n[Done - Tokens: %d]\n", outputTokens)
			break
		}
		fmt.Print(chunk.Text())
	}

	return nil
}
