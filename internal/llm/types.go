package llm

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model       string
	Messages    []Message
	Temperature *float32
	MaxTokens   *int
	Stream      bool
	SessionID   string
}

// StreamOutput describes the behavior consumers need from a streamed model
// response. Each model client can return its own concrete output types.
type StreamOutput interface {
	Text() string
	Done() bool
	TokenUsage() (input, output int)
	Err() error
}
