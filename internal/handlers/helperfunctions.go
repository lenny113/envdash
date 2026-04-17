package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

func GetAndHashAPIKey(r *http.Request) string {
	apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
	if apiKey == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(sum[:])
}
