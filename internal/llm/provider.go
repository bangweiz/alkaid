package llm

import "context"

type Provider interface {
	Stream(ctx context.Context, req Request) (<-chan StreamEvent, error)
}
