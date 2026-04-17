package handlers

import (
	"assignment-2/internal/store"
	"assignment-2/internal/utils"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// DashboardHandler routes incoming HTTP requests to the appropriate handler
// based on the HTTP method. Currently only GET is supported.
func (h *Handler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.DashboardsGetHandler(w, r)
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// DashboardsGetHandler handles GET requests for a specific dashboard.
// It authenticates the request via API key, retrieves the associated country
// registration, fetches aggregated data from cache, and returns a formatted
// dashboard JSON response.
func (h *Handler) DashboardsGetHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the dashboard ID from the URL path
	id := getIDFromDashPath(r.URL.Path)

	// Extract and hash the API key from the request headers
	apiKey := GetAndHashAPIKey(r)

	// Verify that the hashed API key exists in the store
	if !h.store.ApiKeyExists(r.Context(), apiKey) {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Retrieve the country registration associated with the API key and dashboard ID
	registration, err := h.store.GetRegistration(r.Context(), apiKey, id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	// Build a cache request from the registration's country and feature flags
	req := store.CacheExternalRequest{
		Name: registration.Country,
		CCA2: registration.IsoCode,

		// Always fetch country name and ISO code
		CountryName: true,
		CountryCCA2: true,

		// Conditionally fetch weather and air quality data based on feature flags
		MeanTemperature:   registration.Features.Temperature,
		MeanPrecipitation: registration.Features.Precipitation,
		MeanPM25:          registration.Features.AirQuality,
		MeanPM10:          registration.Features.AirQuality,

		// Conditionally fetch geographical and demographic data
		Capital:     registration.Features.Capital,
		Coordinates: registration.Features.Coordinates,
		Population:  registration.Features.Population,
		Area:        registration.Features.Area,

		// Conditionally fetch currency exchange rates
		CurrencyRates: registration.Features.TargetCurrencies,
	}

	// Fetch data from cache (or external sources if not cached)
	cacheResp, err := h.cache.RequestFromCache(req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Assemble the dashboard response payload with all available features
	dashboard := map[string]interface{}{
		"country": safeString(cacheResp.CountryName),
		"isoCode": safeString(cacheResp.CountryCCA2),
		"features": map[string]interface{}{
			"temperature":      cacheResp.MeanTemperature,
			"precipitation":    cacheResp.MeanPrecipitation,
			"capital":          cacheResp.Capital,
			"coordinates":      cacheResp.Coordinates,
			"population":       cacheResp.Population,
			"area":             cacheResp.Area,
			"targetCurrencies": cacheResp.CurrencyRates,
		},
		// Timestamp of when the dashboard data was last retrieved
		"lastRetrieval": time.Now().Format("20060102 15:04"),
	}

	// Append air quality data only if at least one PM metric is available
	if cacheResp.MeanPM25 != nil || cacheResp.MeanPM10 != nil {
		dashboard["features"].(map[string]interface{})["airQuality"] = map[string]interface{}{
			"pm25":  safeFloat(cacheResp.MeanPM25),
			"pm10":  safeFloat(cacheResp.MeanPM10),
			"level": calculateAQLevel(cacheResp.MeanPM25),
		}
	}

	enc := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json") // Set the content type of response
	w.WriteHeader(http.StatusOK)                       // Set HTTP status code to 200 OK

	// Pretty-print JSON output with 2-space indentation
	enc.SetIndent("", "  ")
	// Encode and write the dashboard map to the response body
	enc.Encode(dashboard)
}

// getIDFromDashPath extracts the registration ID from the URL path
// by stripping the known dashboard prefix and any surrounding slashes.
func getIDFromDashPath(path string) string {
	// Trim the dashboard path prefix, leaving only the ID segment
	id := strings.TrimPrefix(path, utils.DASHBOARD_PATH)
	// Remove any leading or trailing slashes from the remaining segment
	return strings.Trim(id, "/")
}

// safeString dereferences a string pointer, returning an empty string if nil.
// Used to safely handle optional string fields from cache responses.
func safeString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// safeFloat dereferences a float64 pointer, returning -1 if nil.
// Used to safely handle optional numeric fields (e.g. air quality metrics).
func safeFloat(ptr *float64) float64 {
	if ptr == nil {
		return -1
	}
	return *ptr
}

// calculateAQLevel maps a PM2.5 concentration (µg/m³) to a human-readable
// air quality level based on EPA breakpoints. Returns "unknown" if input is nil.
func calculateAQLevel(pm25 *float64) string {
	if pm25 == nil {
		return "unknown"
	}

	val := *pm25

	switch {
	case val <= 12:
		return "good"
	case val <= 35:
		return "moderate"
	case val <= 55:
		return "unhealthy for sensitive groups"
	case val <= 150:
		return "unhealthy"
	default:
		return "hazardous"
	}
}
