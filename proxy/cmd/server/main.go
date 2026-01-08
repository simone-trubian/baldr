package main

import (
	"log"
	"net/http"
	"os"

	"github.com/simone-trubian/baldr/proxy/internal/adapters"
	"github.com/simone-trubian/baldr/proxy/internal/core"
	"github.com/simone-trubian/baldr/proxy/internal/handlers"
)

func main() {
	// 1. Configuration (Env Vars)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. Dependency Injection (Wiring)
	// TODO Swap these lines for the real adapters
	llmProvider := &adapters.MockLLM{}
	guardrailService := &adapters.MockGuardrail{}

	// 3. Service Initialization
	svc := core.NewProxyService(llmProvider, guardrailService)
	handler := handlers.NewProxyHandler(svc)

	// 4. Router Setup
	mux := http.NewServeMux()
	mux.HandleFunc("POST /generate", handler.Generate)

	// 5. Start Server
	log.Printf("🛡️ Baldr Proxy running on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}