package handlers

import (
	"assignment-2/internal/models"
	"assignment-2/internal/store"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const dashBase = "/envdash/v1/dashboards/"

// --- Mock Cache ---

type mockCache struct {
	response *store.CacheResponse
	err      error
}

func newTestHandler(store *store.MockStore, cache *mockCache) *Handler {
	return &Handler{store: store, cache: cache}
}

func (m *mockCache) RequestFromCache(_ store.CacheExternalRequest) (*store.CacheResponse, error) {
	return m.response, m.err
}

// --- Helpers ---

func strPtr(s string) *string { return &s }

func floatPtr(f float64) *float64 { return &f }

func seedRegistration(t *testing.T, ms *store.MockStore, reg models.Registration) string {
	t.Helper()
	id, err := ms.CreateRegistration(context.TODO(), "valid", reg)
	if err != nil {
		t.Fatalf("seedRegistration: %v", err)
	}
	return id
}

// --- Router tests ---

func TestDashboardHandler_MethodNotAllowed(t *testing.T) {
	h := newTestHandler(store.NewMockStore(), &mockCache{})

	for _, method := range []string{
		http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch,
	} {
		req := httptest.NewRequest(method, dashBase+"test-id", nil)
		w := httptest.NewRecorder()
		h.DashboardHandler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, w.Code)
		}
	}
}

func TestDashboardHandler_GetRouted(t *testing.T) {
	h := newTestHandler(store.NewMockStore(), &mockCache{})

	req := httptest.NewRequest(http.MethodGet, dashBase+"test-id", nil)
	w := httptest.NewRecorder()
	h.DashboardHandler(w, req)

	if w.Code == http.StatusMethodNotAllowed {
		t.Error("GET should not return 405")
	}
}

// --- Authentication tests ---

func TestDashboardsGetHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(store.NewMockStore(), &mockCache{})

	req := httptest.NewRequest(http.MethodGet, dashBase+"test-id", nil)
	w := httptest.NewRecorder()
	h.DashboardsGetHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// --- Registration lookup tests ---

func TestDashboardsGetHandler_RegistrationNotFound(t *testing.T) {
	ms := store.ValidStore()
	h := newTestHandler(ms, &mockCache{})

	req := httptest.NewRequest(http.MethodGet, dashBase+"nonexistent", nil)
	req.Header.Set("X-API-Key", "valid")
	w := httptest.NewRecorder()
	h.DashboardsGetHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// --- Cache error test ---

func TestDashboardsGetHandler_CacheError(t *testing.T) {
	ms := store.ValidStore()
	seedRegistration(t, ms, models.Registration{Country: "Norway", IsoCode: "NO"})

	c := &mockCache{err: errors.New("cache unavailable")}
	h := newTestHandler(ms, c)

	req := httptest.NewRequest(http.MethodGet, dashBase+"test-id", nil)
	req.Header.Set("X-API-Key", "valid")
	w := httptest.NewRecorder()
	h.DashboardsGetHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// --- Successful response tests ---

func TestDashboardsGetHandler_Success(t *testing.T) {
	ms := store.ValidStore()
	seedRegistration(t, ms, models.Registration{Country: "Norway", IsoCode: "NO"})

	c := &mockCache{
		response: &store.CacheResponse{
			CountryName: strPtr("Norway"),
			CountryCCA2: strPtr("NO"),
		},
	}
	h := newTestHandler(ms, c)

	req := httptest.NewRequest(http.MethodGet, dashBase+"test-id", nil)
	req.Header.Set("X-API-Key", "valid")
	w := httptest.NewRecorder()
	h.DashboardsGetHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["country"] != "Norway" {
		t.Errorf("country: expected %q, got %v", "Norway", body["country"])
	}
	if body["isoCode"] != "NO" {
		t.Errorf("isoCode: expected %q, got %v", "NO", body["isoCode"])
	}
	if _, ok := body["lastRetrieval"]; !ok {
		t.Error("missing lastRetrieval field")
	}
}

func TestDashboardsGetHandler_ContentTypeJSON(t *testing.T) {
	ms := store.ValidStore()
	seedRegistration(t, ms, models.Registration{Country: "Norway", IsoCode: "NO"})

	c := &mockCache{
		response: &store.CacheResponse{
			CountryName: strPtr("Norway"),
			CountryCCA2: strPtr("NO"),
		},
	}
	h := newTestHandler(ms, c)

	req := httptest.NewRequest(http.MethodGet, dashBase+"test-id", nil)
	req.Header.Set("X-API-Key", "valid")
	w := httptest.NewRecorder()
	h.DashboardsGetHandler(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
	}
}

// --- Air quality tests ---

func TestDashboardsGetHandler_AirQualityIncluded(t *testing.T) {
	ms := store.ValidStore()
	seedRegistration(t, ms, models.Registration{
		Country:  "Norway",
		IsoCode:  "NO",
		Features: models.Features{AirQuality: true},
	})

	c := &mockCache{
		response: &store.CacheResponse{
			CountryName: strPtr("Norway"),
			CountryCCA2: strPtr("NO"),
			MeanPM25:    floatPtr(10.0),
			MeanPM10:    floatPtr(20.0),
		},
	}
	h := newTestHandler(ms, c)

	req := httptest.NewRequest(http.MethodGet, dashBase+"test-id", nil)
	req.Header.Set("X-API-Key", "valid")
	w := httptest.NewRecorder()
	h.DashboardsGetHandler(w, req)

	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)

	features := body["features"].(map[string]interface{})
	aq, ok := features["airQuality"]
	if !ok {
		t.Fatal("expected airQuality in features")
	}
	aqMap := aq.(map[string]interface{})
	if aqMap["level"] != "good" {
		t.Errorf("level: expected %q, got %v", "good", aqMap["level"])
	}
}

func TestDashboardsGetHandler_AirQualityOmitted(t *testing.T) {
	ms := store.ValidStore()
	seedRegistration(t, ms, models.Registration{Country: "Norway", IsoCode: "NO"})

	c := &mockCache{
		response: &store.CacheResponse{
			CountryName: strPtr("Norway"),
			CountryCCA2: strPtr("NO"),
		},
	}
	h := newTestHandler(ms, c)

	req := httptest.NewRequest(http.MethodGet, dashBase+"test-id", nil)
	req.Header.Set("X-API-Key", "valid")
	w := httptest.NewRecorder()
	h.DashboardsGetHandler(w, req)

	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)

	features := body["features"].(map[string]interface{})
	if _, ok := features["airQuality"]; ok {
		t.Error("airQuality should be absent when no PM data is available")
	}
}

func TestDashboardsGetHandler_OnlyPM25(t *testing.T) {
	ms := store.ValidStore()
	seedRegistration(t, ms, models.Registration{Country: "Norway", IsoCode: "NO"})

	c := &mockCache{
		response: &store.CacheResponse{
			CountryName: strPtr("Norway"),
			CountryCCA2: strPtr("NO"),
			MeanPM25:    floatPtr(5.0),
		},
	}
	h := newTestHandler(ms, c)

	req := httptest.NewRequest(http.MethodGet, dashBase+"test-id", nil)
	req.Header.Set("X-API-Key", "valid")
	w := httptest.NewRecorder()
	h.DashboardsGetHandler(w, req)

	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)

	features := body["features"].(map[string]interface{})
	if _, ok := features["airQuality"]; !ok {
		t.Error("airQuality should be present when at least PM2.5 is available")
	}
}

// --- Unit tests for helper functions ---

func TestGetIDFromDashPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/envdash/v1/dashboards/abc123", "abc123"},
		{"/envdash/v1/dashboards/abc123/", "abc123"},
		{"/envdash/v1/dashboards/", ""},
	}
	for _, tt := range tests {
		got := getIDFromDashPath(tt.path)
		if got != tt.expected {
			t.Errorf("getIDFromDashPath(%q) = %q, want %q", tt.path, got, tt.expected)
		}
	}
}

func TestSafeString(t *testing.T) {
	if safeString(nil) != "" {
		t.Error("safeString(nil) should return empty string")
	}
	s := "hello"
	if safeString(&s) != "hello" {
		t.Errorf("safeString(&s) = %q, want %q", safeString(&s), "hello")
	}
}

func TestSafeFloat(t *testing.T) {
	if safeFloat(nil) != -1 {
		t.Error("safeFloat(nil) should return -1")
	}
	f := 3.14
	if safeFloat(&f) != 3.14 {
		t.Errorf("safeFloat(&f) = %v, want 3.14", safeFloat(&f))
	}
}

func TestCalculateAQLevel(t *testing.T) {
	tests := []struct {
		pm25     *float64
		expected string
	}{
		{nil, "unknown"},
		{floatPtr(0), "good"},
		{floatPtr(12), "good"},
		{floatPtr(13), "moderate"},
		{floatPtr(35), "moderate"},
		{floatPtr(36), "unhealthy for sensitive groups"},
		{floatPtr(55), "unhealthy for sensitive groups"},
		{floatPtr(56), "unhealthy"},
		{floatPtr(150), "unhealthy"},
		{floatPtr(151), "hazardous"},
		{floatPtr(500), "hazardous"},
	}
	for _, tt := range tests {
		got := calculateAQLevel(tt.pm25)
		if got != tt.expected {
			t.Errorf("calculateAQLevel(%v) = %q, want %q", tt.pm25, got, tt.expected)
		}
	}
}
