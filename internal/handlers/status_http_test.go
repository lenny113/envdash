//go:build flaky
// +build flaky

package handlers

import (
	currencyclient "assignment-2/internal/client/currency"
	aqclient "assignment-2/internal/client/openaq"
	weatherclient "assignment-2/internal/client/openmeteo"
	countryclient "assignment-2/internal/client/restcountries"
	utils "assignment-2/internal/utils"
	"os"
	"strings"
	"testing"
	"time"
)

func newTestStatusHandler(t *testing.T) *StatusHandler {
	t.Helper()

	httpClient := utils.NewHttpClient()
	openAQAPIKey := strings.TrimSpace(os.Getenv("OPENAQ_API_KEY"))

	return NewStatusHandler(
		countryclient.NewRestCountriesClient(httpClient),
		weatherclient.NewWeatherClient(httpClient),
		aqclient.NewOpenAQClient(httpClient, openAQAPIKey),
		currencyclient.NewCurrencyClient(httpClient),
		nil,
		time.Now().Add(-10*time.Second),
	)
}
