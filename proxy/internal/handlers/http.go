package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/simone-trubian/baldr/proxy/internal/core"
)

type ProxyHandler struct {
	service *core.ProxyService
}

func NewProxyHandler(service *core.ProxyService) *ProxyHandler {
	return &ProxyHandler{service: service}
}

func (h *ProxyHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var payload core.RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response, err := h.service.HandleRequest(r.Context(), payload)
	if err != nil {
		// In a real app, check error type to distinguish 400 vs 500
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": response})
}
