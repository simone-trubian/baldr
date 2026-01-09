package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/simone-trubian/baldr/proxy/internal/core"
)

type HTTPHandler struct {
	service core.ProxyServicePort
}

func NewHTTPHandler(s core.ProxyServicePort) *HTTPHandler {
	return &HTTPHandler{service: s}
}

func (h *HTTPHandler) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload core.RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Context propagation is automatic here
	response, err := h.service.Execute(r.Context(), payload)
	if err != nil {
		// TODO differentiate between 400 and 500 errors
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": response})
}
