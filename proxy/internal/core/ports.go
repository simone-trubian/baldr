package core

import (
	"context"
)

// Domain Models
type RequestPayload struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type GuardrailResponse struct {
	Allowed        bool   `json:"allowed"`
	Reason         string `json:"reason,omitempty"`
	SanitizedInput string `json:"sanitized_input,omitempty"`
}

// --- Ports (Interfaces) ---

// GuardrailPort defines the contract for guardrail service.
type GuardrailPort interface {
	Validate(ctx context.Context, input string) (GuardrailResponse, error)
}

// LLMPort defines the contract for external AI providers.
type LLMPort interface {
	Generate(ctx context.Context, payload RequestPayload) (string, error)
}

// ProxyServicePort defines the main entry point for the business logic.
type ProxyServicePort interface {
	Execute(ctx context.Context, payload RequestPayload) (string, error)
}
