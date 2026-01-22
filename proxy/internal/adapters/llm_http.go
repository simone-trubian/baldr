package adapters

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LLMConfig struct {
	BaseURL string
	APIKey  string
}

type LLM struct {
	client     *http.Client
	baseURL    string
	targetHost string
	apiKey     string
}

func NewLLM(config LLMConfig) *LLM {
	return &LLM{
		client: &http.Client{
			Timeout: 60 * time.Second, // Long timeout for LLM generation
		},
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
	}
}

func (a *LLM) Generate(ctx context.Context, payload []byte, headers map[string]string) (io.ReadCloser, error) {

	// Re-create the request for the upstream
	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	// Copy headers (important for Auth/Content-Type)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Set the REAL upstream key (from config)
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		// Clean up the body if we aren't returning it
		resp.Body.Close()
		return nil, fmt.Errorf("upstream returned status: %d", resp.StatusCode)
	}

	// Return the body directly for streaming
	return resp.Body, nil
}
