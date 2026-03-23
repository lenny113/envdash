package utils

import (
	"net/http"
	"time"
)

func NewHttpClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
	}
}
