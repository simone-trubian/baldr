package ports

import (
	"context"

	"github.com/simone-trubian/baldr/proxy/internal/core/domain"
)

// GuardrailPort defines the contract for guardrail service.
type GuardrailPort interface {
	// It returns the decision, reason, and potentially modified input.
	Validate(ctx context.Context, payload []byte) (*domain.GuardrailResponse, error)
}
