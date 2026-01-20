package ports

import (
	"context"
	"io"
)

// ProxyServicePort defines the main entry point for the business logic.
type ProxyServicePort interface {
	Execute(ctx context.Context, payload []byte, headers map[string]string) (io.ReadCloser, error)
}
