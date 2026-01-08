package adapters

import (
	"context"
	"github.com/simone-trubian/baldr/proxy/internal/core"
	"log"
	"time"
)

// MockLLM simulates OpenAI without the bill
type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, payload core.RequestPayload) (string, error) {
	// Simulate network latency
	time.Sleep(500 * time.Millisecond)
	log.Printf("[MockLLM] Generating response for model: %s", payload.Model)
	return "This is a deterministic response from Baldr Mock Provider.", nil
}

// MockGuardrail simulates the Python service
type MockGuardrail struct{}

func (m *MockGuardrail) ScanPrompt(ctx context.Context, prompt string) (bool, error) {
	// Simulate PII check latency
	time.Sleep(50 * time.Millisecond)
	if prompt == "INJECT_ATTACK" {
		return false, nil
	}
	return true, nil
}