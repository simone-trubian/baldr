package ports

import (
	"context"

	"github.com/simone-trubian/baldr/proxy/internal/core/domain"
)

// LLMPort defines the contract for external AI providers.
type LLMPort interface {
	Generate(ctx context.Context, payload domain.RequestPayload) (string, error)
}
