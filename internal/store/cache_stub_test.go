//go:build !flaky

package store

import (
	currencyclient "assignment-2/internal/client/currency"
	aqclient "assignment-2/internal/client/openaq"
	weatherclient "assignment-2/internal/client/openmeteo"
	countryclient "assignment-2/internal/client/restcountries"
	"testing"
)

type stubCountryClient struct {
	resp countryclient.RestCountries_INT_Response
	err  error
}

func (s *stubCountryClient) GetCountryInfo(req countryclient.RestCountries_InformationRequest) (countryclient.RestCountries_INT_Response, error) {
	return s.resp, s.err
}

type stubWeatherClient struct {
	resp weatherclient.Weather_INT_Response
	err  error
}

func (s *stubWeatherClient) GetInfo(req weatherclient.Weather_InformationRequest) (weatherclient.Weather_INT_Response, error) {
	return s.resp, s.err
}

type stubCurrencyClient struct {
	resp currencyclient.Currency_INT_Response
	err  error
}

func (s *stubCurrencyClient) GetSelectedExchangeRates(baseCurrency string) (currencyclient.Currency_INT_Response, error) {
	return s.resp, s.err
}

type stubAQClient struct {
	resp aqclient.OpenAQ_INT_Response
	err  error
}

func (s *stubAQClient) GetInfo(req aqclient.OpenAQ_InformationRequest) (aqclient.OpenAQ_INT_Response, error) {
	return s.resp, s.err
}

func strPtr(s string) *string       { return &s }
func int64Ptr(v int64) *int64       { return &v }
func float64Ptr(v float64) *float64 { return &v }

func newTestCache(t *testing.T) *Cache {
	t.Helper()

	countryResp := countryclient.RestCountries_INT_Response{
		Country:     strPtr("Norway"),
		CCA2:        strPtr("NO"),
		Capital:     strPtr("Oslo"),
		Coordinates: &[]float64{62, 10},
		Population:  int64Ptr(5379475),
		Area:        float64Ptr(323802),
		Borders:     &[]string{"FIN", "SWE", "RUS"},
		Currencies:  &[]string{"NOK"},
	}

	meanTemp := 2.5
	meanPrecip := 0.1
	weatherResp := weatherclient.Weather_INT_Response{
		MeanTemperature:   &meanTemp,
		MeanPrecipitation: &meanPrecip,
	}

	currencyResp := currencyclient.Currency_INT_Response{
		BaseCurrency: "NOK",
		Rates: map[string]float64{
			"EUR": 0.09,
			"SEK": 0.97,
		},
	}

	meanPM25 := 7.5
	meanPM10 := 12.0
	aqResp := aqclient.OpenAQ_INT_Response{
		MeanPM25: &meanPM25,
		MeanPM10: &meanPM10,
	}

	return NewCache(
		&stubCountryClient{resp: countryResp},
		&stubWeatherClient{resp: weatherResp},
		&stubCurrencyClient{resp: currencyResp},
		&stubAQClient{resp: aqResp},
	)
}
