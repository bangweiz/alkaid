package agents

import (
	"context"
	"fmt"

	"github.com/bangweiz/alkaid/internal/llm"
	"github.com/bangweiz/alkaid/internal/llm/gemini"
)

type CodingAgent struct {
	provider llm.Provider
}

func NewCodingAgent(p llm.Provider) *CodingAgent {
	return &CodingAgent{provider: p}
}

func (a *CodingAgent) Prompt(ctx context.Context, text string) error {
	req := llm.Request{
		Model: string(gemini.ModelGemini35FlashLite),
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "You are a Go specialist."},
			{Role: llm.RoleUser, Content: text},
		},
	}

	stream, err := a.provider.Stream(ctx, req)
	if err != nil {
		return err
	}

	for chunk := range stream {
		if chunk.Error != nil {
			return chunk.Error
		}
		if chunk.IsDone {
			fmt.Printf("\n[Done - Tokens: %d]\n", chunk.OutputTokens)
			break
		}
		fmt.Print(chunk.Text)
	}

	return nil
}
