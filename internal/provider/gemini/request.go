package gemini

type Model string

const (
	ModelGemini38Flash     Model = "gemini-3.8-flash"
	ModelGemini35FlashLite Model = "gemini-3.5-flash-lite"
)

type ThinkingLevel string

const (
	ThinkingLevelNone   ThinkingLevel = "none"
	ThinkingLevelLow    ThinkingLevel = "low"
	ThinkingLevelMedium ThinkingLevel = "medium"
	ThinkingLevelHigh   ThinkingLevel = "high"
)

type InteractionRequest struct {
	Model                 Model             `json:"model"`
	SystemInstruction     string            `json:"system_instruction,omitempty"`
	Input                 string            `json:"input"`
	GenerationConfig      *GenerationConfig `json:"generation_config,omitempty"`
	Stream                bool              `json:"stream,omitempty"`
	PreviousInteractionID string            `json:"previous_interaction_id,omitempty"`
}

type GenerationConfig struct {
	ThinkingLevel ThinkingLevel `json:"thinking_level,omitempty"`
}

func NewInteractionRequest(model Model, input string) InteractionRequest {
	req := InteractionRequest{
		Model: model,
		Input: input,
	}

	return req
}
