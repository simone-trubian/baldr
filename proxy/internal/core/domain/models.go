package domain

import "encoding/json"

// RequestPayload represents the core input to the system.
type RequestPayload struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	// TODO implement stream observing Stream bool   `json:"stream"`
}

type GuardrailResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
	// Use RawMessage so we can capture any JSON structure (dict, list, etc.)
	SanitizedInput json.RawMessage `json:"sanitized_input,omitempty"`
}
