package utils

import (
	"net/http"
	"time"
)

func newHttpClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
	}
}
