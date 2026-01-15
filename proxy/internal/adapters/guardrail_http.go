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

type RemoteGuardrail struct {
	client  *http.Client
	baseURL string
}

func NewRemoteGuardrail(url string) *RemoteGuardrail {
	return &RemoteGuardrail{
		client:  &http.Client{Timeout: 2 * time.Second}, // Fail fast!
		baseURL: url,
	}
}

type guardrailRequest struct {
	Prompt string `json:"prompt"`
}

type guardrailResponse struct {
	Allowed        bool   `json:"allowed"`
	Reason         string `json:"reason"`
	SanitizedInput string `json:"sanitized_input"`
}

func (r *RemoteGuardrail) Validate(ctx context.Context, input string) (domain.GuardrailResponse, error) {
	reqBody, _ := json.Marshal(guardrailRequest{Prompt: input})

	req, err := http.NewRequestWithContext(ctx, "POST", r.baseURL+"/validate", bytes.NewBuffer(reqBody))
	if err != nil {
		return domain.GuardrailResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return domain.GuardrailResponse{}, fmt.Errorf("guardrail service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.GuardrailResponse{}, fmt.Errorf("guardrail returned status: %d", resp.StatusCode)
	}

	var res guardrailResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return domain.GuardrailResponse{}, err
	}

	return domain.GuardrailResponse{
		Allowed:        res.Allowed,
		Reason:         res.Reason,
		SanitizedInput: res.SanitizedInput,
	}, nil
}
