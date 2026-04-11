package client

import (
	"testing"
)

func TestBuildURL_ValidBaseCurrency(t *testing.T) {
	req := Currency_InformationRequest{
		BaseCurrency: NOK,
		Currencies:   []CurrencyCode{EUR, SEK},
	}

	fullURL, err := buildURL(req)
	if err != nil {
		t.Fatal(err)
	}

	expected := base_url + "NOK"
	if fullURL != expected {
		t.Fatalf("expected URL %q, got %q", expected, fullURL)
	}
}

func TestBuildURL_MissingBaseCurrency(t *testing.T) {
	req := Currency_InformationRequest{
		Currencies: []CurrencyCode{EUR, SEK},
	}

	_, err := buildURL(req)
	if err == nil {
		t.Fatal("expected error when base currency is missing")
	}
}

func TestDecodeResponse_Valid(t *testing.T) {
	body := []byte(`{
		"result": "success",
		"base_code": "NOK",
		"rates": {
			"EUR": 0.089044,
			"SEK": 0.969592,
			"USD": 0.10273
		}
	}`)

	data, err := decodeResponse(body)
	if err != nil {
		t.Fatal(err)
	}

	if data.BaseCode != "NOK" {
		t.Fatalf("expected base_code to be NOK, got %q", data.BaseCode)
	}

	if len(data.Rates) != 3 {
		t.Fatalf("expected 3 rates, got %d", len(data.Rates))
	}

	if data.Rates["EUR"] != 0.089044 {
		t.Errorf("expected EUR rate 0.089044, got %v", data.Rates["EUR"])
	}

	if data.Rates["SEK"] != 0.969592 {
		t.Errorf("expected SEK rate 0.969592, got %v", data.Rates["SEK"])
	}

	if data.Rates["USD"] != 0.10273 {
		t.Errorf("expected USD rate 0.10273, got %v", data.Rates["USD"])
	}
}

func TestDecodeResponse_MissingBase(t *testing.T) {
	body := []byte(`{
		"result": "success",
		"rates": {
			"EUR": 0.089044
		}
	}`)

	_, err := decodeResponse(body)
	if err == nil {
		t.Fatal("expected error when base_code is missing")
	}
}

func TestDecodeResponse_MissingRates(t *testing.T) {
	body := []byte(`{
		"result": "success",
		"base_code": "NOK"
	}`)

	_, err := decodeResponse(body)
	if err == nil {
		t.Fatal("expected error when rates are missing")
	}
}

func TestGetSelectedExchangeRates_FiltersRequestedCurrencies(t *testing.T) {
	req := Currency_InformationRequest{
		BaseCurrency: NOK,
		Currencies:   []CurrencyCode{EUR, SEK},
	}

	fullURL, err := buildURL(req)
	if err != nil {
		t.Fatal(err)
	}
	if fullURL != base_url+"NOK" {
		t.Fatalf("unexpected URL: %q", fullURL)
	}

	decoded, err := decodeResponse([]byte(`{
		"result": "success",
		"base_code": "NOK",
		"rates": {
			"EUR": 0.089044,
			"SEK": 0.969592,
			"USD": 0.10273
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}

	filteredRates := make(map[CurrencyCode]float64)
	for _, currency := range req.Currencies {
		rate, exists := decoded.Rates[string(currency)]
		if exists {
			filteredRates[currency] = rate
		}
	}

	response := Currency_INT_Response{
		BaseCurrency: CurrencyCode(decoded.BaseCode),
		Rates:        filteredRates,
	}

	if response.BaseCurrency != NOK {
		t.Fatalf("expected base currency NOK, got %q", response.BaseCurrency)
	}

	if len(response.Rates) != 2 {
		t.Fatalf("expected 2 filtered rates, got %d", len(response.Rates))
	}

	if response.Rates[EUR] != 0.089044 {
		t.Errorf("expected EUR rate 0.089044, got %v", response.Rates[EUR])
	}

	if response.Rates[SEK] != 0.969592 {
		t.Errorf("expected SEK rate 0.969592, got %v", response.Rates[SEK])
	}

	if _, exists := response.Rates[USD]; exists {
		t.Error("did not expect USD in filtered response")
	}
}
