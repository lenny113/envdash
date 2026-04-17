//go:build flaky
// +build flaky

package handlers

import (
	currencyclient "assignment-2/internal/client/currency"
	aqclient "assignment-2/internal/client/openaq"
	weatherclient "assignment-2/internal/client/openmeteo"
	countryclient "assignment-2/internal/client/restcountries"
	utils "assignment-2/internal/utils"
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TODO: Replace mockStatusStore with a real Firestore integration setup for flaky tests.
// For now, flaky tests use real external HTTP API calls and a mocked Firestore store
// so the shared status tests can pass without local Firebase credentials.
type mockStatusStore struct{}

func (m *mockStatusStore) DB_Status(ctx context.Context) bool {
	return true
}

func (m *mockStatusStore) CountFirestore(ctx context.Context, collection string) (int, error) {
	return 0, nil
}

func newTestStatusHandler(t *testing.T) *StatusHandler {
	t.Helper()

	httpClient := utils.NewHttpClient()
	openAQAPIKey := strings.TrimSpace(os.Getenv("OPENAQ_API_KEY"))

	return NewStatusHandler(
		countryclient.NewRestCountriesClient(httpClient),
		weatherclient.NewWeatherClient(httpClient),
		aqclient.NewOpenAQClient(httpClient, openAQAPIKey),
		currencyclient.NewCurrencyClient(httpClient),
		&mockStatusStore{},
		time.Now().Add(-10*time.Second),
	)
}
