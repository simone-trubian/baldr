package core

import (
	"context"
	"fmt"
	"log"
)

type BaldrService struct {
	guardrail GuardrailPort
	llm       LLMPort
}

// NewBaldrService injects dependencies.
// This is critical for testability: we pass interfaces, not concrete types.
func NewBaldrService(g GuardrailPort, l LLMPort) *BaldrService {
	return &BaldrService{
		guardrail: g,
		llm:       l,
	}
}

func (s *BaldrService) Execute(ctx context.Context, payload RequestPayload) (string, error) {
	// 1. Guardrail Check (Orchestrator Pattern)
	check, err := s.guardrail.Validate(ctx, payload.Prompt)
	if err != nil {
		// Devil's Advocate: Fail Closed!
		// If the guardrail is down, we must NOT let traffic through.
		return "", fmt.Errorf("safety check system failure: %w", err)
	}

	if !check.Allowed {
		log.Printf("[Block] Prompt blocked. Reason: %s", check.Reason)
		return "", fmt.Errorf("safety violation: %s", check.Reason)
	}

	// 2. Use Sanitized Input (if provided)
	if check.SanitizedInput != "" {
		payload.Prompt = check.SanitizedInput
	}

	// 3. Call LLM
	response, err := s.llm.Generate(ctx, payload)
	if err != nil {
		return "", fmt.Errorf("upstream provider error: %w", err)
	}

	return response, nil
}
