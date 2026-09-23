package llm

import "context"

type Client interface {
	Stream(context.Context, Request) (<-chan StreamOutput, error)
}
