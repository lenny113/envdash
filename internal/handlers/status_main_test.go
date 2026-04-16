package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetStatus_ReturnsStatusResponse(t *testing.T) {
	h := newTestStatusHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/envdash/v1/status", nil)
	rr := httptest.NewRecorder()

	h.GetStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", rr.Code)
	}

	var resp StatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if resp.CountriesAPI != http.StatusOK {
		t.Fatalf("expected countries_api 200, got %d", resp.CountriesAPI)
	}
	if resp.MeteoAPI != http.StatusOK {
		t.Fatalf("expected meteo_api 200, got %d", resp.MeteoAPI)
	}
	if resp.OpenAQAPI != http.StatusOK {
		t.Fatalf("expected openaq_api 200, got %d", resp.OpenAQAPI)
	}
	if resp.CurrencyAPI != http.StatusOK {
		t.Fatalf("expected currency_api 200, got %d", resp.CurrencyAPI)
	}
	if resp.NotificationDB != http.StatusOK {
		t.Fatalf("expected notification_db 200, got %d", resp.NotificationDB)
	}
	if resp.Webhooks != 0 {
		t.Fatalf("expected webhooks 0, got %d", resp.Webhooks)
	}
	if resp.Version != "v1" {
		t.Fatalf("expected version v1, got %q", resp.Version)
	}
	if resp.Uptime < 0 {
		t.Fatalf("expected uptime >= 0, got %d", resp.Uptime)
	}
}

func TestGetStatus_UsesCachedResponseWithinRefreshWindow(t *testing.T) {
	h := newTestStatusHandler(t)

	firstReq := httptest.NewRequest(http.MethodGet, "/envdash/v1/status", nil)
	firstRR := httptest.NewRecorder()
	h.GetStatus(firstRR, firstReq)

	if h.cached == nil {
		t.Fatal("expected cached response to be stored after first request")
	}

	cachedBefore := *h.cached
	lastRefreshBefore := h.lastRefresh

	time.Sleep(100 * time.Millisecond)

	secondReq := httptest.NewRequest(http.MethodGet, "/envdash/v1/status", nil)
	secondRR := httptest.NewRecorder()
	h.GetStatus(secondRR, secondReq)

	if secondRR.Code != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", secondRR.Code)
	}

	var resp StatusResponse
	if err := json.Unmarshal(secondRR.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if h.lastRefresh != lastRefreshBefore {
		t.Fatal("expected cached response to be reused without refreshing")
	}

	if resp.CountriesAPI != cachedBefore.CountriesAPI {
		t.Fatalf("expected cached countries_api %d, got %d", cachedBefore.CountriesAPI, resp.CountriesAPI)
	}
	if resp.MeteoAPI != cachedBefore.MeteoAPI {
		t.Fatalf("expected cached meteo_api %d, got %d", cachedBefore.MeteoAPI, resp.MeteoAPI)
	}
	if resp.OpenAQAPI != cachedBefore.OpenAQAPI {
		t.Fatalf("expected cached openaq_api %d, got %d", cachedBefore.OpenAQAPI, resp.OpenAQAPI)
	}
	if resp.CurrencyAPI != cachedBefore.CurrencyAPI {
		t.Fatalf("expected cached currency_api %d, got %d", cachedBefore.CurrencyAPI, resp.CurrencyAPI)
	}
	if resp.Version != cachedBefore.Version {
		t.Fatalf("expected cached version %q, got %q", cachedBefore.Version, resp.Version)
	}
}

func TestGetStatus_RefreshesAfterWindowExpires(t *testing.T) {
	h := newTestStatusHandler(t)

	h.cached = &StatusResponse{
		CountriesAPI:   http.StatusInternalServerError,
		MeteoAPI:       http.StatusInternalServerError,
		OpenAQAPI:      http.StatusInternalServerError,
		CurrencyAPI:    http.StatusInternalServerError,
		NotificationDB: http.StatusOK,
		Webhooks:       0,
		Version:        "v1",
		Uptime:         1,
	}
	h.lastRefresh = time.Now().Add(-10 * time.Second)

	req := httptest.NewRequest(http.MethodGet, "/envdash/v1/status", nil)
	rr := httptest.NewRecorder()

	h.GetStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", rr.Code)
	}

	var resp StatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if resp.CountriesAPI != http.StatusOK {
		t.Fatalf("expected refreshed countries_api 200, got %d", resp.CountriesAPI)
	}
	if resp.MeteoAPI != http.StatusOK {
		t.Fatalf("expected refreshed meteo_api 200, got %d", resp.MeteoAPI)
	}
	if resp.OpenAQAPI != http.StatusOK {
		t.Fatalf("expected refreshed openaq_api 200, got %d", resp.OpenAQAPI)
	}
	if resp.CurrencyAPI != http.StatusOK {
		t.Fatalf("expected refreshed currency_api 200, got %d", resp.CurrencyAPI)
	}
}
