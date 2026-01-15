package ports

import (
	"context"

	"github.com/simone-trubian/baldr/proxy/internal/core/domain"
)

// GuardrailPort defines the contract for guardrail service.
type GuardrailPort interface {
	Validate(ctx context.Context, input string) (domain.GuardrailResponse, error)
}
