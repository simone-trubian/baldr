package core

import (
	"context"

	"github.com/simone-trubian/baldr/proxy/internal/core/domain"
)

// --- Ports (Interfaces) ---

// GuardrailPort defines the contract for guardrail service.
type GuardrailPort interface {
	Validate(ctx context.Context, input string) (domain.GuardrailResponse, error)
}

// LLMPort defines the contract for external AI providers.
type LLMPort interface {
	Generate(ctx context.Context, payload domain.RequestPayload) (string, error)
}

// ProxyServicePort defines the main entry point for the business logic.
type ProxyServicePort interface {
	Execute(ctx context.Context, payload domain.RequestPayload) (string, error)
}
