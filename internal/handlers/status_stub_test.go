//go:build !flaky

package handlers

import (
	currencyclient "assignment-2/internal/client/currency"
	aqclient "assignment-2/internal/client/openaq"
	weatherclient "assignment-2/internal/client/openmeteo"
	countryclient "assignment-2/internal/client/restcountries"
	"testing"
	"time"
)

type stubCountryClient struct {
	err error
}

func (s *stubCountryClient) GetCountryInfo(req countryclient.RestCountries_InformationRequest) (countryclient.RestCountries_INT_Response, error) {
	return countryclient.RestCountries_INT_Response{}, s.err
}

type stubWeatherClient struct {
	err error
}

func (s *stubWeatherClient) GetInfo(req weatherclient.Weather_InformationRequest) (weatherclient.Weather_INT_Response, error) {
	return weatherclient.Weather_INT_Response{}, s.err
}

type stubAQClient struct {
	err error
}

func (s *stubAQClient) GetInfo(req aqclient.OpenAQ_InformationRequest) (aqclient.OpenAQ_INT_Response, error) {
	return aqclient.OpenAQ_INT_Response{}, s.err
}

type stubCurrencyClient struct {
	err error
}

func (s *stubCurrencyClient) GetSelectedExchangeRates(baseCurrency string) (currencyclient.Currency_INT_Response, error) {
	return currencyclient.Currency_INT_Response{}, s.err
}

func newTestStatusHandler(t *testing.T) *StatusHandler {
	t.Helper()

	return NewStatusHandler(
		&stubCountryClient{},
		&stubWeatherClient{},
		&stubAQClient{},
		&stubCurrencyClient{},
		nil,
		time.Now().Add(-10*time.Second),
	)
}
