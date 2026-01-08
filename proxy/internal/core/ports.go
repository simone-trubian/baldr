package core

import (
	"context"
)

// RequestPayload represents the incoming user request
type RequestPayload struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"` // e.g., "gpt-4", "claude-3"
}

// LLMProvider defines the behavior for external AI providers.
// This interface allows us to inject OpenAI, Anthropic, or a Mock.
type LLMProvider interface {
	Generate(ctx context.Context, payload RequestPayload) (string, error)
}

// GuardrailService defines how we talk to the Python sidecar.
// For Day 0, we can mock this too.
type GuardrailService interface {
	ScanPrompt(ctx context.Context, prompt string) (bool, error) // Returns true if safe
}