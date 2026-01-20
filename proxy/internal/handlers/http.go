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

	// headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// The Flushing Loop
	flusher, ok := w.(http.Flusher)
	if !ok {
		// Fallback for non-flushing writers (unlikely in standard net/http)
		io.Copy(w, respStream)
		return
	}

	// Copy buffer and flush immediately
	// Using a small buffer (e.g., 32 bytes) or simply copying byte-by-byte is inefficient.
	// Better: use io.Copy but running in a loop is hard.
	// Simplest robust way for LLM streaming:

	buf := make([]byte, 1024) // 1KB buffer
	for {
		n, err := respStream.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			flusher.Flush() // Force data to client immediately
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error streaming response:", err)
			break
		}
	}
}
