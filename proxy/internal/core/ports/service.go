package ports

import (
	"context"

	"github.com/simone-trubian/baldr/proxy/internal/core/domain"
)

// ProxyServicePort defines the main entry point for the business logic.
type ProxyServicePort interface {
	Execute(ctx context.Context, payload domain.RequestPayload) (string, error)
}
