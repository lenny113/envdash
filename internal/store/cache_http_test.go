//go:build flaky
// +build flaky

package store

import (
	currencyclient "assignment-2/internal/client/currency"
	aqclient "assignment-2/internal/client/openaq"
	weatherclient "assignment-2/internal/client/openmeteo"
	countryclient "assignment-2/internal/client/restcountries"
	utils "assignment-2/internal/utils"
	"os"
	"testing"
)

func newTestCache(t *testing.T) *Cache {
	t.Helper()

	httpClient := utils.NewHttpClient()

	apiKey := os.Getenv("OPENAQ_API_KEY")
	if apiKey == "" {
		t.Skip("skipping flaky test: OPENAQ_API_KEY is not set")
	}

	return NewCache(
		countryclient.NewRestCountriesClient(httpClient),
		weatherclient.NewWeatherClient(httpClient),
		currencyclient.NewCurrencyClient(httpClient),
		aqclient.NewOpenAQClient(httpClient, apiKey),
	)
}
