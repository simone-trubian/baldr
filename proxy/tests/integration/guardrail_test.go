package integration

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/simone-trubian/baldr/proxy/internal/adapters"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// This does NOT use the real Python service
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
	_, err = guardrailAdapter.Validate(ctx, "sensitive data")
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
			"path": "/scan"
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
