package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/simone-trubian/baldr/proxy/internal/core/domain"
)

type GuardrailConfig struct {
	BaseURL string
	Timeout time.Duration
}

type RemoteGuardrail struct {
	client  *http.Client
	baseURL string
}

func NewRemoteGuardrail(config GuardrailConfig) *RemoteGuardrail {
	return &RemoteGuardrail{
		client:  &http.Client{Timeout: config.Timeout}, // Fail fast!
		baseURL: config.BaseURL,
	}
}

type guardrailRequest struct {
	Prompt string `json:"prompt"`
}

type guardrailResponse struct {
	Allowed        bool            `json:"allowed"`
	Reason         string          `json:"reason"`
	SanitizedInput json.RawMessage `json:"sanitized_input,omitempty"`
}

func (a *RemoteGuardrail) Validate(ctx context.Context, payload []byte) (*domain.GuardrailResponse, error) {
	// 1. Prepare Request to Python Sidecar
	// We send the payload exactly as we received it from the user.
	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create guardrail request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 2. Execute
	resp, err := a.client.Do(req)
	if err != nil {
		// This handles timeouts (context deadline) and connection refused
		return nil, fmt.Errorf("guardrail connection error: %w", err)
	}
	defer resp.Body.Close()

	// 3. Handle non-200 codes from the Sidecar logic
	// If the Sidecar crashes (500), we treat it as an error to trigger Fail Closed.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("guardrail sidecar returned status: %d", resp.StatusCode)
	}

	// 4. Decode Response
	var result domain.GuardrailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode guardrail response: %w", err)
	}

	return &result, nil
}
