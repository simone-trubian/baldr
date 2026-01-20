package ports

import (
	"context"
	"io"
)

type LLMPort interface {
	// SendRequest forwards the validated payload to the LLM.
	// Returns a stream (io.ReadCloser) to support SSE, or an error.
	Generate(ctx context.Context, payload []byte, headers map[string]string) (io.ReadCloser, error)
}
