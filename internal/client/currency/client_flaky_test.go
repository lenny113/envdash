//go:build flaky
// +build flaky

package client

import (
	"net/http"
	"testing"
	"time"
)

func TestGetSelectedExchangeRates(t *testing.T) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	currencyClient := NewCurrencyClient(httpClient)

	data, err := currencyClient.GetSelectedExchangeRates("NOK")
	if err != nil {
		t.Fatal(err)
	}

	if data.BaseCurrency != "NOK" {
		t.Fatalf("expected base currency to be NOK, got %q", data.BaseCurrency)
	}

	if len(data.Rates) == 0 {
		t.Fatal("expected rates in response, got none")
	}

	nokRate, ok := data.Rates["NOK"]
	if !ok {
		t.Fatal("expected NOK in response")
	}
	if nokRate != 1 {
		t.Fatalf("expected NOK rate to be 1, got %v", nokRate)
	}

	eurRate, ok := data.Rates["EUR"]
	if !ok {
		t.Fatal("expected EUR in response")
	}
	if eurRate <= 0 {
		t.Errorf("expected EUR rate to be > 0, got %v", eurRate)
	}

	sekRate, ok := data.Rates["SEK"]
	if !ok {
		t.Fatal("expected SEK in response")
	}
	if sekRate <= 0 {
		t.Errorf("expected SEK rate to be > 0, got %v", sekRate)
	}

	usdRate, ok := data.Rates["USD"]
	if !ok {
		t.Fatal("expected USD in response")
	}
	if usdRate <= 0 {
		t.Errorf("expected USD rate to be > 0, got %v", usdRate)
	}

	t.Logf("base currency: %s", data.BaseCurrency)
	t.Logf("number of rates returned: %d", len(data.Rates))
	t.Logf("full conversion map from the Application Programming Interface (API): %#v", data.Rates)
}

func TestGetSelectedExchangeRates_WithTrimmedLowercaseInput(t *testing.T) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	currencyClient := NewCurrencyClient(httpClient)

	data, err := currencyClient.GetSelectedExchangeRates(" nok ")
	if err != nil {
		t.Fatal(err)
	}

	if data.BaseCurrency != "NOK" {
		t.Fatalf("expected base currency to be NOK, got %q", data.BaseCurrency)
	}

	if len(data.Rates) == 0 {
		t.Fatal("expected rates in response, got none")
	}

	t.Logf("full conversion map from the Application Programming Interface (API): %#v", data.Rates)
}
