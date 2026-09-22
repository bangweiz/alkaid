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

type Response struct {
	Text         string
	SessionID    string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

type StreamEvent struct {
	Text         string
	IsDone       bool
	InputTokens  int
	OutputTokens int
	Error        error
}
