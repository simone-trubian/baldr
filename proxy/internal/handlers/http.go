package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/simone-trubian/baldr/proxy/internal/core/ports"
)

type HTTPHandler struct {
	service ports.ProxyServicePort
}

func NewHTTPHandler(s ports.ProxyServicePort) *HTTPHandler {
	return &HTTPHandler{service: s}
}
func (h *HTTPHandler) HandleProxy(w http.ResponseWriter, r *http.Request) {
	// 1. Buffer the body (We need it for both Guardrail and LLM)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	r.Body.Close() // Close the original reader

	// Extract headers to forward
	headers := make(map[string]string)
	headers["Content-Type"] = r.Header.Get("Content-Type")
	headers["Authorization"] = r.Header.Get("Authorization")

	// 2. Call Service
	respStream, err := h.service.Execute(r.Context(), body, headers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	defer respStream.Close()

	// 3. Stream Response back to client
	w.Header().Set("Content-Type", "application/json") // Or text/event-stream
	if _, err := io.Copy(w, respStream); err != nil {
		fmt.Println("Error streaming response:", err)
	}
}
