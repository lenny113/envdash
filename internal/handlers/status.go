package handlers

import (
	currencyclient "assignment-2/internal/client/currency"
	aqclient "assignment-2/internal/client/openaq"
	weatherclient "assignment-2/internal/client/openmeteo"
	countryclient "assignment-2/internal/client/restcountries"
	utils "assignment-2/internal/utils"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type StatusStore interface {
	DB_Status(ctc context.Context) bool
	CountFirestore(ctx context.Context, collection string) (int, error)
}

type StatusHandler struct {
	countryClient  countryclient.RestCountriesClient
	weatherClient  weatherclient.WeatherClient
	aqClient       aqclient.OpenAQClient
	currencyClient currencyclient.CurrencyClient
	store          StatusStore
	startedAt      time.Time

	mu           sync.Mutex
	cached       *StatusResponse
	lastRefresh  time.Time
	refreshEvery time.Duration
}

type StatusResponse struct {
	CountriesAPI   int    `json:"countries_api"`
	MeteoAPI       int    `json:"meteo_api"`
	OpenAQAPI      int    `json:"openaq_api"`
	CurrencyAPI    int    `json:"currency_api"`
	NotificationDB int    `json:"notification_db"`
	Webhooks       int    `json:"webhooks"`
	Version        string `json:"version"`
	Uptime         int64  `json:"uptime"`
}

func NewStatusHandler(
	countryClient countryclient.RestCountriesClient,
	weatherClient weatherclient.WeatherClient,
	aqClient aqclient.OpenAQClient,
	currencyClient currencyclient.CurrencyClient,
	store StatusStore,
	startedAt time.Time,
) *StatusHandler {
	return &StatusHandler{
		countryClient:  countryClient,
		weatherClient:  weatherClient,
		aqClient:       aqClient,
		currencyClient: currencyClient,
		store:          store,
		startedAt:      startedAt,
		refreshEvery:   5 * time.Second,
	}
}

func (h *StatusHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.mu.Lock()
	if h.cached != nil && time.Since(h.lastRefresh) < h.refreshEvery {
		resp := *h.cached
		resp.Uptime = int64(time.Since(h.startedAt).Seconds())
		h.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	h.mu.Unlock()

	resp := StatusResponse{
		CountriesAPI:   probeCountriesAPI(h.countryClient),
		MeteoAPI:       probeMeteoAPI(h.weatherClient),
		OpenAQAPI:      probeOpenAQAPI(h.aqClient),
		CurrencyAPI:    probeCurrencyAPI(h.currencyClient),
		NotificationDB: h.probeNotificationDB(ctx),
		Webhooks:       h.count(ctx, "all_notifications"),
		Version:        utils.VERSION,
		Uptime:         int64(time.Since(h.startedAt).Seconds()),
	}

	h.mu.Lock()
	h.cached = &resp
	h.lastRefresh = time.Now()
	h.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func probeCountriesAPI(client countryclient.RestCountriesClient) int {
	if client == nil {
		return http.StatusInternalServerError
	}

	_, err := client.GetCountryInfo(countryclient.RestCountries_InformationRequest{
		ISOCode: "NO",
		CCA2:    true,
	})
	if err != nil {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func probeMeteoAPI(client weatherclient.WeatherClient) int {
	if client == nil {
		return http.StatusInternalServerError
	}

	_, err := client.GetInfo(weatherclient.Weather_InformationRequest{
		Lat:         62,
		Lng:         10,
		Temperature: true,
	})
	if err != nil {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func probeOpenAQAPI(client aqclient.OpenAQClient) int {
	if client == nil {
		return http.StatusInternalServerError
	}

	_, err := client.GetInfo(aqclient.OpenAQ_InformationRequest{
		ISOCode: "NO",
		PM25:    true,
	})
	if err != nil {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func probeCurrencyAPI(client currencyclient.CurrencyClient) int {
	if client == nil {
		return http.StatusInternalServerError
	}

	_, err := client.GetSelectedExchangeRates("NOK")
	if err != nil {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func (h *StatusHandler) probeNotificationDB(ctx context.Context) int {
	if h.store == nil {
		return http.StatusInternalServerError
	}

	if !h.store.DB_Status(ctx) {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (h *StatusHandler) count(ctx context.Context, collection string) int {

	count, err := h.store.CountFirestore(ctx, collection)
	if err != nil {
		return 0
	}
	return count
}
