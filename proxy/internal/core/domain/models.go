package domain

// RequestPayload represents the core input to the system.
type RequestPayload struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	// TODO implement stream observing Stream bool   `json:"stream"`
}

// GuardrailResponse captures the safety check result.
type GuardrailResponse struct {
	Allowed        bool   `json:"allowed"`
	Reason         string `json:"reason,omitempty"`
	SanitizedInput string `json:"sanitized_input,omitempty"`
}
