package gemini

import (
	"encoding/json"
	"fmt"
)

func UnmarshalStreamEvent(raw []byte) (any, error) {
	var header StreamEventHeader
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, fmt.Errorf("gemini: failed to extract event_type: %w", err)
	}

	switch header.EventType {
	case EventInteractionCreated:
		var ev InteractionCreatedEvent
		return &ev, json.Unmarshal(raw, &ev)

	case EventInteractionStatusUpdate:
		var ev InteractionStatusUpdateEvent
		return &ev, json.Unmarshal(raw, &ev)

	case EventStepStart:
		var ev StepStartEvent
		return &ev, json.Unmarshal(raw, &ev)

	case EventStepDelta:
		var ev StepDeltaEvent
		return &ev, json.Unmarshal(raw, &ev)

	case EventStepStop:
		var ev StepStopEvent
		return &ev, json.Unmarshal(raw, &ev)

	case EventInteractionCompleted:
		var ev InteractionCompletedEvent
		return &ev, json.Unmarshal(raw, &ev)

	case EventError:
		var ev StreamErrorEvent
		return &ev, json.Unmarshal(raw, &ev)

	default:
		return nil, fmt.Errorf("gemini: unknown event_type '%s'", header.EventType)
	}
}
