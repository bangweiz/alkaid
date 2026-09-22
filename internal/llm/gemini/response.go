package gemini

import "time"

type StreamEventType string

const (
	EventInteractionCreated      StreamEventType = "interaction.created"
	EventInteractionStatusUpdate StreamEventType = "interaction.status_update"
	EventStepStart               StreamEventType = "step.start"
	EventStepDelta               StreamEventType = "step.delta"
	EventStepStop                StreamEventType = "step.stop"
	EventInteractionCompleted    StreamEventType = "interaction.completed"
	EventError                   StreamEventType = "error"
)

type StreamEventHeader struct {
	EventType StreamEventType `json:"event_type"`
}

type InteractionCreatedEvent struct {
	EventType   StreamEventType            `json:"event_type"`
	Interaction InteractionCreatedMetadata `json:"interaction"`
}

type InteractionCreatedMetadata struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Object string `json:"object"`
	Model  Model  `json:"model"`
}

type InteractionStatusUpdateEvent struct {
	EventType     StreamEventType `json:"event_type"`
	InteractionID string          `json:"interaction_id"`
	Status        string          `json:"status"`
}

type StepStartEvent struct {
	EventType StreamEventType `json:"event_type"`
	Index     int             `json:"index"`
	Step      StepStartDetail `json:"step"`
}

type StepStartDetail struct {
	Type      string         `json:"type"`
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type StepDeltaEvent struct {
	EventType StreamEventType `json:"event_type"`
	Index     int             `json:"index"`
	Delta     StepDelta       `json:"delta"`
}

type StepDelta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`

	Signature string        `json:"signature,omitempty"`
	Content   *DeltaContent `json:"content,omitempty"`

	Arguments string `json:"arguments,omitempty"`

	MimeType string `json:"mime_type,omitempty"`
	Data     string `json:"data,omitempty"`
}

type DeltaContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type StepStopEvent struct {
	EventType StreamEventType `json:"event_type"`
	Index     int             `json:"index"`
	Usage     *StreamUsage    `json:"usage,omitempty"`      // Running total usage
	StepUsage *StreamUsage    `json:"step_usage,omitempty"` // Usage for this step alone
}

type InteractionCompletedEvent struct {
	EventType   StreamEventType            `json:"event_type"`
	Interaction CompletedInteractionDetail `json:"interaction"`
}

type CompletedInteractionDetail struct {
	ID          string      `json:"id"`
	Status      string      `json:"status"`
	Object      string      `json:"object"`
	Model       Model       `json:"model"`
	ServiceTier string      `json:"service_tier,omitempty"`
	Created     time.Time   `json:"created,omitempty"`
	Updated     time.Time   `json:"updated,omitempty"`
	Usage       StreamUsage `json:"usage"`
}

type StreamUsage struct {
	TotalTokens           int                     `json:"total_tokens"`
	TotalInputTokens      int                     `json:"total_input_tokens"`
	TotalOutputTokens     int                     `json:"total_output_tokens"`
	TotalCachedTokens     int                     `json:"total_cached_tokens,omitempty"`
	TotalThoughtTokens    int                     `json:"total_thought_tokens,omitempty"`
	TotalToolUseTokens    int                     `json:"total_tool_use_tokens,omitempty"`
	InputTokensByModality []InputTokensByModality `json:"input_tokens_by_modality,omitempty"`
}

type InputTokensByModality struct {
	Modality string `json:"modality"`
	Tokens   int    `json:"tokens"`
}

type StreamErrorEvent struct {
	EventType StreamEventType `json:"event_type"`
	Error     StreamErrorData `json:"error"`
}

type StreamErrorData struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
