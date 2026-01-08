package core

import (
	"context"
	"fmt"
)

type ProxyService struct {
	llm       LLMProvider
	guardrail GuardrailService
}

func NewProxyService(llm LLMProvider, guardrail GuardrailService) *ProxyService {
	return &ProxyService{
		llm:       llm,
		guardrail: guardrail,
	}
}

func (s *ProxyService) HandleRequest(ctx context.Context, payload RequestPayload) (string, error) {
	// 1. Guardrail Check (Synchronous)
	safe, err := s.guardrail.ScanPrompt(ctx, payload.Prompt)
	if err != nil {
		return "", fmt.Errorf("guardrail error: %w", err)
	}
	if !safe {
		return "", fmt.Errorf("security violation: prompt rejected by Baldr")
	}

	// 2. LLM Generation
	// TODO add Async Logging here (fire and forget)
	response, err := s.llm.Generate(ctx, payload)
	if err != nil {
		return "", fmt.Errorf("provider error: %w", err)
	}

	return response, nil
}