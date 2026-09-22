package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/bangweiz/alkaid/internal/llm/gemini"
)

type GeminiAdapter struct {
	client *gemini.Client
}

func NewGeminiAdapter(client *gemini.Client) *GeminiAdapter {
	return &GeminiAdapter{client: client}
}

func (a *GeminiAdapter) Stream(ctx context.Context, req Request) (<-chan StreamEvent, error) {
	geminiReq := a.toGeminiRequest(req)
	geminiReq.Stream = true

	eventChan, err := a.client.CreateInteraction(ctx, &geminiReq)
	if err != nil {
		return nil, fmt.Errorf("gemini adapter stream: %w", err)
	}

	outChan := make(chan StreamEvent)

	go func() {
		defer close(outChan)

		for res := range eventChan {
			if res.Error != nil {
				outChan <- StreamEvent{Error: res.Error}
				return
			}

			switch ev := res.Event.(type) {
			case *gemini.StepDeltaEvent:
				if ev.Delta.Type == "text" {
					outChan <- StreamEvent{Text: ev.Delta.Text}
				}

			case *gemini.InteractionCompletedEvent:
				outChan <- StreamEvent{
					IsDone:       true,
					InputTokens:  ev.Interaction.Usage.TotalInputTokens,
					OutputTokens: ev.Interaction.Usage.TotalOutputTokens,
				}

			case *gemini.StreamErrorEvent:
				outChan <- StreamEvent{
					Error: fmt.Errorf("[%s] %s", ev.Error.Code, ev.Error.Message),
				}
				return
			}
		}
	}()

	return outChan, nil
}

func (a *GeminiAdapter) toGeminiRequest(req Request) gemini.InteractionRequest {
	var systemInstruction string
	var userInput strings.Builder

	for _, m := range req.Messages {
		if m.Role == RoleSystem {
			systemInstruction = m.Content
		} else {
			userInput.WriteString(m.Content + "\n")
		}
	}

	return gemini.InteractionRequest{
		Model:                 gemini.Model(req.Model),
		SystemInstruction:     systemInstruction,
		Input:                 strings.TrimSpace(userInput.String()),
		PreviousInteractionID: req.SessionID,
		Stream:                true,
	}
}
