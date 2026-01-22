package core

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/simone-trubian/baldr/proxy/internal/core/ports"
)

type BaldrService struct {
	guardrail ports.GuardrailPort
	llm       ports.LLMPort
}

func NewBaldrService(g ports.GuardrailPort, l ports.LLMPort) *BaldrService {
	return &BaldrService{
		guardrail: g,
		llm:       l,
	}
}

// Orchestration method
func (s *BaldrService) Execute(ctx context.Context, payload []byte, headers map[string]string) (io.ReadCloser, error) {
	// 1. Guardrail Check
	decision, err := s.guardrail.Validate(ctx, payload)
	if err != nil {
		// FAIL CLOSED: Any technical error blocks the request.
		return nil, fmt.Errorf("guardrail check failed (fail-closed): %w", err)
	}

	// 2. Policy Enforcement
	if !decision.Allowed {
		return nil, fmt.Errorf("blocked: %s", decision.Reason)
	}

	// 3. PII Redaction / Sanitization Logic
	// If the Python sidecar redacted data, we MUST use the new payload.
	finalPayload := payload
	if !bytes.Equal(decision.SanitizedInput, []byte("null")) {
		// Assuming SanitizedInput is the full JSON body string.
		// Convert back to bytes.
		finalPayload = []byte(decision.SanitizedInput)
	}

	// 4. Upstream to LLM using finalPayload
	responseStream, err := s.llm.Generate(ctx, finalPayload, headers)
	if err != nil {
		return nil, fmt.Errorf("upstream llm error: %w", err)
	}

	return responseStream, nil
}
