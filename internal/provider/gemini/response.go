package gemini

import "time"

type InteractionResponse struct {
	ID      string    `json:"id"`
	Status  string    `json:"status"`
	Usage   Usage     `json:"usage"`
	Created time.Time `json:"created"`
	Steps   []Step    `json:"steps"`
	Object  string    `json:"object"`
	Model   string    `json:"model"`
}

type Usage struct {
	TotalTokens       int `json:"total_tokens"`
	TotalInputTokens  int `json:"total_input_tokens"`
	TotalOutputTokens int `json:"total_output_tokens"`
}

type Step struct {
	Type      string    `json:"type"`
	Signature string    `json:"signature,omitempty"`
	Content   []Content `json:"content,omitempty"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
