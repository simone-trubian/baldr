package main

import (
	"log"
	"net/http"

	"github.com/simone-trubian/baldr/proxy/internal/adapters"
	"github.com/simone-trubian/baldr/proxy/internal/core"
	"github.com/simone-trubian/baldr/proxy/internal/handlers"
)

func main() {
	// 1. Initialize Adapters (Infrastructure)
	// Later, these will be RealGuardrailAdapter and OpenAIAdapter
	guardrail := &adapters.MockGuardrail{}
	llm := &adapters.MockLLM{}

	// 2. Initialize Service (Core Logic)
	// Dependency Injection happens here
	service := core.NewBaldrService(guardrail, llm)

	// 3. Initialize Handlers (Presentation)
	handler := handlers.NewHTTPHandler(service)

	// 4. Start Server
	mux := http.NewServeMux()
	mux.HandleFunc("/generate", handler.HandleGenerate)

	log.Println("Baldr Proxy Service running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
