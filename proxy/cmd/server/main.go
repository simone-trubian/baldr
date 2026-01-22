package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/simone-trubian/baldr/proxy/internal/adapters"
	"github.com/simone-trubian/baldr/proxy/internal/core"
	"github.com/simone-trubian/baldr/proxy/internal/handlers"
)

// Simple config loader
type Config struct {
	ServerPort           string
	GuardrailURL         string
	LLMURL               string
	LLMAPIKey            string
	GuardrailConcurrency int
	GuarailTimeout       int
}

func loadConfig() Config {
	return Config{
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		GuardrailURL:         getEnv("GUARDRAIL_URL", "http://localhost:8000/validate"), // Default to local sidecar
		LLMURL:               getEnv("LLM_URL", "https://generativelanguage.googleapis.com/v1beta/openai/"),
		LLMAPIKey:            getEnv("LLM_API_KEY", ""),
		GuardrailConcurrency: getEnvInt("GUARDRAIL_MAX_CONCURRENCY", 50), // Default 50 concurrent checks
		GuarailTimeout:       getEnvInt("GUARDRAIL_TIMEOUT", 1),
	}
}

func main() {
	// 1. Configuration
	cfg := loadConfig()
	log.Printf("Starting Baldr Proxy on port %s", cfg.ServerPort)
	log.Printf("Guardrail: %s (Concurrency Limit: %d)", cfg.GuardrailURL, cfg.GuardrailConcurrency)
	log.Printf("Upstream LLM: %s", cfg.LLMURL)

	guardrailConfig := adapters.GuardrailConfig{
		BaseURL:        cfg.GuardrailURL,
		Timeout:        time.Duration(cfg.GuarailTimeout) * time.Second,
		MaxConcurrency: cfg.GuardrailConcurrency,
	}
	llmConfig := adapters.LLMConfig{
		BaseURL: cfg.LLMURL,
		APIKey:  cfg.LLMAPIKey,
	}
	guardrailAdapter := adapters.NewRemoteGuardrail(guardrailConfig)
	llmAdapter := adapters.NewLLM(llmConfig)

	// 2. Initialize Service (Core Logic)
	// Dependency Injection happens here
	service := core.NewBaldrService(guardrailAdapter, llmAdapter)

	// 3. Initialize Handlers (Presentation)
	handler := handlers.NewHTTPHandler(service)

	// 4. Router Setup
	mux := http.NewServeMux()
	// Map the proxy endpoint. You might want to make the path configurable too.
	mux.HandleFunc("POST /chat/completions", handler.HandleProxy)

	// Health check for Docker/K8s
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 5. Server Configuration
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,  // Time to read the incoming request body
		WriteTimeout: 0,                 // Must be 0 (infinite) for LLM Streaming!
		IdleTimeout:  120 * time.Second, // Keep-alive connections
	}

	// 6. Graceful Shutdown Routine
	// We want to handle SIGINT (Ctrl+C) and SIGTERM (Docker stop)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for current requests to complete
	// Give existing LLM streams 30 seconds to finish before killing them.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
