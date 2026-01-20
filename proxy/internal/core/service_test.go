package core_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/simone-trubian/baldr/proxy/internal/core"
	"github.com/simone-trubian/baldr/proxy/internal/core/domain"
)

// Test Mocks defined locally to control behavior per test
type TestMockGuardrail struct {
	mockValidate func(ctx context.Context, payload []byte) (*domain.GuardrailResponse, error)
}

func (a *TestMockGuardrail) Validate(ctx context.Context, payload []byte) (*domain.GuardrailResponse, error) {
	return a.mockValidate(ctx, payload)
}

type TestMockLLM struct {
	mockGenerate func(ctx context.Context, payload []byte, headers map[string]string) (io.ReadCloser, error)
}

func (m *TestMockLLM) Generate(ctx context.Context, payload []byte, headers map[string]string) (io.ReadCloser, error) {
	return m.mockGenerate(ctx, payload, headers)
}

func TestBaldrService_FailClosed(t *testing.T) {
	// Scenario: The Guardrail service is down (returns error)
	// Expected: Service should return error, LLM should NOT be called.

	guardrail := &TestMockGuardrail{
		mockValidate: func(ctx context.Context, payload []byte) (*domain.GuardrailResponse, error) {
			return &domain.GuardrailResponse{}, errors.New("guardrail unreachable")
		},
	}
	llm := &TestMockLLM{
		mockGenerate: func(ctx context.Context, payload []byte, headers map[string]string) (io.ReadCloser, error) {
			t.Fatal("LLM should not be called if guardrail fails")
			return nil, nil
		},
	}

	service := core.NewBaldrService(guardrail, llm)
	_, err := service.Execute(context.Background(), []byte{}, make(map[string]string))

	if err == nil {
		t.Error("Expected error due to guardrail failure, got nil")
	}
}

func TestBaldrService_Sanitization(t *testing.T) {
	// Scenario: Guardrail flags PII and returns sanitized prompt.
	// Expected: LLM receives the SANITIZED prompt, not original.

	guardrail := &TestMockGuardrail{
		mockValidate: func(ctx context.Context, payload []byte) (*domain.GuardrailResponse, error) {
			return &domain.GuardrailResponse{Allowed: true, SanitizedInput: []byte("safe prompt")}, nil
		},
	}
	llm := &TestMockLLM{
		mockGenerate: func(ctx context.Context, payload []byte, headers map[string]string) (io.ReadCloser, error) {
			if !bytes.Equal(payload, []byte("safe prompt")) {
				t.Errorf("Expected LLM to receive 'safe prompt', got '%s'", payload)
			}
			stringReader := strings.NewReader("ok")
			return io.NopCloser(stringReader), nil
		},
	}

	service := core.NewBaldrService(guardrail, llm)
	service.Execute(
		context.Background(), []byte("unsafe prompt"), make(map[string]string))
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
