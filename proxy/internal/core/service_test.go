package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/simone-trubian/baldr/proxy/internal/core"
	"github.com/simone-trubian/baldr/proxy/internal/core/domain"
)

// Test Mocks defined locally to control behavior per test
type TestMockGuardrail struct {
	mockValidate func(ctx context.Context, input string) (domain.GuardrailResponse, error)
}

func (m *TestMockGuardrail) Validate(ctx context.Context, input string) (domain.GuardrailResponse, error) {
	return m.mockValidate(ctx, input)
}

type TestMockLLM struct {
	mockGenerate func(ctx context.Context, payload domain.RequestPayload) (string, error)
}

func (m *TestMockLLM) Generate(ctx context.Context, payload domain.RequestPayload) (string, error) {
	return m.mockGenerate(ctx, payload)
}

func TestBaldrService_FailClosed(t *testing.T) {
	// Scenario: The Guardrail service is down (returns error)
	// Expected: Service should return error, LLM should NOT be called.

	guardrail := &TestMockGuardrail{
		mockValidate: func(ctx context.Context, input string) (domain.GuardrailResponse, error) {
			return domain.GuardrailResponse{}, errors.New("guardrail unreachable")
		},
	}
	llm := &TestMockLLM{
		mockGenerate: func(ctx context.Context, payload domain.RequestPayload) (string, error) {
			t.Fatal("LLM should not be called if guardrail fails")
			return "", nil
		},
	}

	service := core.NewBaldrService(guardrail, llm)
	_, err := service.Execute(context.Background(), domain.RequestPayload{Prompt: "Hello"})

	if err == nil {
		t.Error("Expected error due to guardrail failure, got nil")
	}
}

func TestBaldrService_Sanitization(t *testing.T) {
	// Scenario: Guardrail flags PII and returns sanitized prompt.
	// Expected: LLM receives the SANITIZED prompt, not original.

	guardrail := &TestMockGuardrail{
		mockValidate: func(ctx context.Context, input string) (domain.GuardrailResponse, error) {
			return domain.GuardrailResponse{Allowed: true, SanitizedInput: "safe prompt"}, nil
		},
	}
	llm := &TestMockLLM{
		mockGenerate: func(ctx context.Context, payload domain.RequestPayload) (string, error) {
			if payload.Prompt != "safe prompt" {
				t.Errorf("Expected LLM to receive 'safe prompt', got '%s'", payload.Prompt)
			}
			return "ok", nil
		},
	}

	service := core.NewBaldrService(guardrail, llm)
	service.Execute(context.Background(), domain.RequestPayload{Prompt: "unsafe prompt"})
}

/*
func TestService_ProcessRequest_FailClosed(t *testing.T) {
	// Define the "Table"
	tests := []struct {
		name           string
		guardrailResp  error // Simulate adapter failure
		failClosedMode bool
		expectedStatus int
	}{
		{"Guardrail Down - Fail Closed", errors.New("timeout"), true, 500},
		{"Guardrail Down - Fail Open", errors.New("timeout"), false, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup logic...
			// Assertions using "github.com/stretchr/testify/assert"
		})
	}
}
*/
