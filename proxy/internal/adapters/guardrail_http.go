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
	BaseURL        string
	Timeout        time.Duration
	MaxConcurrency int
}

type RemoteGuardrail struct {
	client    *http.Client
	baseURL   string
	semaphore chan struct{} // Sets rate limit
}

func NewRemoteGuardrail(config GuardrailConfig) *RemoteGuardrail {
	return &RemoteGuardrail{
		client:    &http.Client{Timeout: config.Timeout},
		baseURL:   config.BaseURL,
		semaphore: make(chan struct{}, config.MaxConcurrency),
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
	// 1. Acquire token
	// If the channel is full, this blocks until a request returns
	select {
	case a.semaphore <- struct{}{}:
	case <-ctx.Done():
		return nil, fmt.Errorf("Request cancelled while awaiting for Guardrail service to become available")
	}
	// 2. Release the token on exit
	defer func() { <-a.semaphore }()

	// 3. Prepare Request to Python Sidecar
	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create guardrail request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 4. Execute
	resp, err := a.client.Do(req)
	if err != nil {
		// This handles timeouts (context deadline) and connection refused
		return nil, fmt.Errorf("guardrail connection error: %w", err)
	}
	defer resp.Body.Close()

	// 5. Handle non-200 codes from the Sidecar logic
	// If the Sidecar crashes (500), we treat it as an error to trigger Fail Closed.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("guardrail sidecar returned status: %d", resp.StatusCode)
	}

	// 6. Decode Response
	var result domain.GuardrailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode guardrail response: %w", err)
	}

	return &result, nil
}
