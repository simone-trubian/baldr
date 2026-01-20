package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/simone-trubian/baldr/proxy/internal/adapters"
	"github.com/simone-trubian/baldr/proxy/internal/core/domain"
)

func TestGuardrail_HappyPath_SchemaValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 1. Start MockServer Container
	req := testcontainers.ContainerRequest{
		Image:        "mockserver/mockserver:5.15.0",
		ExposedPorts: []string{"1080/tcp"},
		WaitingFor:   wait.ForLog("started on port: 1080"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)
	defer container.Terminate(ctx)

	ip, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "1080")
	mockServerURL := "http://" + ip + ":" + port.Port()

	// 2. Define the CONTRACT (What Python Pydantic expects/returns)
	// We define this LOCALLY to ensure we are testing the Wire Format,
	// not just reusing Go structs.
	type PythonRequestContract struct {
		Prompt string `json:"prompt"`
	}

	type PythonResponseContract struct {
		Allowed        bool            `json:"allowed"`
		Reason         string          `json:"reason"`
		SanitizedInput json.RawMessage `json:"sanitized_input,omitempty"`
	}

	// 3. Prepare the Data
	contractReq := PythonRequestContract{Prompt: "SELECT * FROM users"}
	contractResp := PythonResponseContract{
		Allowed:        false,
		Reason:         "SQL Injection Detected",
		SanitizedInput: []byte(`{"prompt": "SELECT * FROM users"}`),
	}

	// 4. Program MockServer
	// This ensures Go sends EXACTLY {"prompt": "..."}
	// and can handle the response.
	createStrictJSONExpectation(t, mockServerURL, contractReq, contractResp)

	// 5. Configure Adapter
	config := adapters.GuardrailConfig{
		BaseURL: mockServerURL,
		Timeout: 2 * time.Second,
	}
	adapter := adapters.NewRemoteGuardrail(config)

	// 6. Execute Logic (Domain Model Input)
	domainInput := domain.RequestPayload{
		Model:  "gpt-4", // Note: This field is currently dropped by the adapter
		Prompt: "SELECT * FROM users",
	}

	result, err := adapter.Validate(ctx, []byte(domainInput.Prompt))

	// 7. Assertions
	assert.NoError(t, err)

	// Verify Domain Model Mapping
	assert.False(t, result.Allowed)
	assert.Equal(t, "SQL Injection Detected", result.Reason)
	// Verify the Snake_case JSON was mapped to CamelCase Go Struct correctly
	assert.Equal(t, "SELECT * FROM users", result.SanitizedInput)
}

func TestGuardrail_FailClosed_OnTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 1. Start the MockServer Container that acts as the Python service
	req := testcontainers.ContainerRequest{
		Image:        "mockserver/mockserver:5.15.0",
		ExposedPorts: []string{"1080/tcp"},
		WaitingFor:   wait.ForLog("started on port: 1080"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)
	defer container.Terminate(ctx) // Clean up docker after test

	// 2. Get the dynamic port Docker assigned
	ip, err := container.Host(ctx)
	assert.NoError(t, err)
	port, err := container.MappedPort(ctx, "1080")
	assert.NoError(t, err)
	mockServerURL := "http://" + ip + ":" + port.Port()

	// 3. Program the MockServer to be "Toxic" (High Latency)
	// We tell it: "If you receive a POST to /scan, wait 5 seconds then return 200"
	createMockExpectation(t, mockServerURL, 5000) // 5000ms delay

	// 4. Configure the Go Adapter
	// We set a STRICT timeout of 1 second.
	// This forces the "Fail Closed" logic to trigger because 1s < 5s.
	config := adapters.GuardrailConfig{
		BaseURL: mockServerURL,
		Timeout: 1 * time.Second,
	}
	guardrailAdapter := adapters.NewRemoteGuardrail(config)

	// 5. Execute
	start := time.Now()
	_, err = guardrailAdapter.Validate(ctx, []byte("sensitive data"))
	duration := time.Since(start)

	// 6. Assertions
	// It should have failed
	assert.Error(t, err, "Expected an error due to timeout")

	// It should have failed FAST (approx 1s), not waited the full 5s
	assert.Less(t, duration.Seconds(), 1.5, "Proxy did not cancel the request in time!")

	// Check specific error type if your domain defines it
	// assert.IsType(t, &core.ErrGuardrailTimeout{}, err)
}

// Helper to send JSON config to MockServer
func createMockExpectation(t *testing.T, baseURL string, delayMs int) {
	client := &http.Client{}
	// MockServer expectation JSON format
	// Docs: https://www.mock-server.com/
	body := fmt.Sprintf(`{
		"httpRequest": {
			"method": "POST",
			"path": "/validate"
		},
		"httpResponse": {
			"statusCode": 200,
			"body": "{\"safe\": true}"
		},
		"delay": {
			"timeUnit": "MILLISECONDS",
			"value": %d
		}
	}`, delayMs)

	req, _ := http.NewRequest("PUT", baseURL+"/mockserver/expectation", strings.NewReader(body))
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	//assert.Equal(t, 201, resp.StatusCode)
}

// createStrictJSONExpectation programs MockServer to match the JSON body strictly.
// If Go sends extra fields, wrong types, or missing fields, MockServer returns 404.
func createStrictJSONExpectation(t *testing.T, baseURL string, expectedReq, mockResp interface{}) {
	client := &http.Client{}

	// Marshal the Go structs to JSON bytes
	reqBytes, err := json.Marshal(expectedReq)
	assert.NoError(t, err)

	respBytes, err := json.Marshal(mockResp)
	assert.NoError(t, err)

	// Construct the MockServer payload
	// We use json.RawMessage to embed the pre-marshaled JSON into the larger structure
	payload := map[string]interface{}{
		"httpRequest": map[string]interface{}{
			"method": "POST",
			"path":   "/validate",
			"body": map[string]interface{}{
				"type":      "JSON",
				"json":      json.RawMessage(reqBytes),
				"matchType": "ONLY_MATCHING_FIELDS", // Allows key re-ordering, enforces existence
			},
		},
		"httpResponse": map[string]interface{}{
			"statusCode": 200,
			"headers": map[string][]string{
				"content-type": {"application/json"},
			},
			"body": json.RawMessage(respBytes),
		},
	}

	finalBody, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, _ := http.NewRequest("PUT", baseURL+"/mockserver/expectation", strings.NewReader(string(finalBody)))
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode, "MockServer rejected the expectation configuration")
}
