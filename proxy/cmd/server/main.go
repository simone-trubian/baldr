package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/simone-trubian/baldr/proxy/internal/adapters"
	"github.com/simone-trubian/baldr/proxy/internal/core"
	"github.com/simone-trubian/baldr/proxy/internal/handlers"
)

func main() {
	// 1. Initialize Adapters (Infrastructure)
	config := adapters.GuardrailConfig{
		BaseURL: os.Getenv("GUARDRAIL_URL"),
		Timeout: 1 * time.Second,
	}
	guardrailAdapter := adapters.NewRemoteGuardrail(config)
	llmAdapter := adapters.NewLLM("/target") // TODO config for LLM adapter

	// 2. Initialize Service (Core Logic)
	// Dependency Injection happens here
	service := core.NewBaldrService(guardrailAdapter, llmAdapter)

	// 3. Initialize Handlers (Presentation)
	handler := handlers.NewHTTPHandler(service)

	// 4. Start Server
	mux := http.NewServeMux()
	mux.HandleFunc("/generate", handler.HandleProxy)

	log.Println("Baldr Proxy Service running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
