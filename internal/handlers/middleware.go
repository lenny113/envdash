package handlers

import (
	"assignment-2/internal/utils"
	"net/http"
	"strings"
)

// AuthMiddleware validates the X-API-Key header on protected routes.
// Returns 401 if key is missing, 403 if key is invalid or revoked.
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))

		if apiKey == "" {
			writeJSONError(w, http.StatusUnauthorized, "Missing API key, include 'X-API-Key' in header")
			utils.SetMessageForLogger(w, "Unauthorized access attempt: missing API key")
			return
		}

		if !h.store.ApiKeyExists(r.Context(), apiKey) {
			writeJSONError(w, http.StatusForbidden, "Invalid API key")
			utils.SetMessageForLogger(w, "Forbidden access attempt: invalid API key")
			return
		}

		next.ServeHTTP(w, r)
	})
}
