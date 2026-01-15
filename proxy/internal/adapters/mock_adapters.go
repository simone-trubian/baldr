package adapters

import (
	"context"
	"strings"
	"time"

	"github.com/simone-trubian/baldr/proxy/internal/core/domain"
)

// MockGuardrail simulates the Python FastAPI service.
type MockGuardrail struct{}

func (m *MockGuardrail) Validate(ctx context.Context, input string) (domain.GuardrailResponse, error) {
	// Simulate Network Latency
	time.Sleep(20 * time.Millisecond)

	if strings.Contains(input, "ATTACK") {
		return domain.GuardrailResponse{
			Allowed: false,
			Reason:  "Malicious injection detected",
		}, nil
	}

	// Example of sanitization (e.g., PII masking)
	sanitized := strings.ReplaceAll(input, "password", "[REDACTED]")

	return domain.GuardrailResponse{
		Allowed:        true,
		SanitizedInput: sanitized,
	}, nil
}

// MockLLM simulates OpenAI/Anthropic.
type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, payload domain.RequestPayload) (string, error) {
	// Simulate Generation Latency
	time.Sleep(100 * time.Millisecond)
	return "This is a response from the Baldr Mock LLM for: " + payload.Prompt, nil
}
