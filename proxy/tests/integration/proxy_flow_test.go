package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/simone-trubian/baldr/proxy/internal/adapters"
	"github.com/simone-trubian/baldr/proxy/internal/core"
	"github.com/simone-trubian/baldr/proxy/internal/handlers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestProxyFlow_Sanitization(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mockserver/mockserver:5.15.0",
		ExposedPorts: []string{"1080/tcp"},
		// Wait for the actual HTTP Health endpoint, not the log.
		// This ensures the port is reachable from Go.
		WaitingFor: wait.ForHTTP("/mockserver/status").
			WithPort("1080/tcp").
			WithMethod("PUT").
			WithStartupTimeout(60 * time.Second),
	}

	mockContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer mockContainer.Terminate(ctx)

	endpoint, _ := mockContainer.Endpoint(ctx, "")
	mockServerBaseUrl := fmt.Sprintf("http://%s", endpoint)

	// --- DEFINE EXPECTATIONS  ---

	// A. Guardrail Expectation: Receives "dirty" input -> Returns "sanitized" input
	configureMockServer(
		t,
		mockServerBaseUrl,
		"/guardrail",
		"POST",
		`{"prompt": "my secret password"}`, // Expected Input
		200,
		`{"allowed": true, "sanitized_input": {"prompt": "my secret [REDACTED]"}}`) // Response

	// B. LLM Expectation: MUST receive "sanitized" input
	configureMockServer(
		t,
		mockServerBaseUrl,
		"/v1/chat/completions",
		"POST",
		`{"prompt": "my secret [REDACTED]"}`, // If it receives "password", this won't match -> 404
		200,
		`{"id": "chatcmpl-123", "choices": [{"message": {"content": "Hello"}}]}`,
	)

	// --- SETUP SYSTEM UNDER TEST ---

	guardrailConfig := adapters.GuardrailConfig{
		BaseURL: mockServerBaseUrl + "/guardrail",
		Timeout: 2 * time.Second,
	}
	llmConfig := adapters.LLMConfig{
		BaseURL: mockServerBaseUrl + "/v1/chat/completions",
		APIKey:  "key",
	}

	guardrailAdapter := adapters.NewRemoteGuardrail(guardrailConfig)
	llmAdapter := adapters.NewLLM(llmConfig)

	svc := core.NewBaldrService(guardrailAdapter, llmAdapter)
	handler := handlers.NewHTTPHandler(svc)

	// --- EXECUTE TEST ---

	// The User sends the "Dirty" request
	reqBody := `{"prompt": "my secret password"}`
	r := httptest.NewRequest("POST", "/proxy", strings.NewReader(reqBody))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleProxy(w, r)

	// --- ASSERTIONS ---

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	// 1. Check Status
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Proxy should succeed")

	// 2. Check that we got the LLM response
	assert.Contains(t, string(body), "chatcmpl-123", "Should receive upstream LLM response")

	// 3. (Implicit)
	// If the Proxy failed to swap the input, the LLM Mock would have rejected
	// the request (because we configured it to ONLY match the redacted body),
	// causing the LLM Adapter to return 404/500, failing assertion #1.
}

// Helper to configure MockServer via HTTP (Standard MockServer API)
func configureMockServer(t *testing.T, host, path, method, bodyMatch string, respCode int, respBody string) {
	payload := fmt.Sprintf(`{
        "httpRequest": {
            "method": "%s",
            "path": "%s",
            "body": {
                "type": "JSON",
                "json": %s
            }
        },
        "httpResponse": {
            "statusCode": %d,
            "body": %s
        }
    }`, method, path, bodyMatch, respCode, respBody)

	// Use a custom client with a short timeout to prevent hangs
	client := &http.Client{Timeout: 5 * time.Second}

	// Ensure the URL is valid
	url := fmt.Sprintf("%s/mockserver/expectation", host)
	req, err := http.NewRequest("PUT", url, strings.NewReader(payload))
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to connect to MockServer to set expectation")
	defer resp.Body.Close() // FIX: Close the body

	require.Equal(t, 201, resp.StatusCode, "MockServer rejected the expectation")
}
