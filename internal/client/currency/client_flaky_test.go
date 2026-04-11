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

	req := Currency_InformationRequest{
		BaseCurrency: NOK,
		Currencies:   []CurrencyCode{EUR, SEK},
	}

	data, err := currencyClient.GetSelectedExchangeRates(req)
	if err != nil {
		t.Fatal(err)
	}

	if data.BaseCurrency != NOK {
		t.Fatalf("expected base currency to be NOK, got %q", data.BaseCurrency)
	}

	if len(data.Rates) != 2 {
		t.Fatalf("expected 2 rates, got %d: %#v", len(data.Rates), data.Rates)
	}

	eurRate, ok := data.Rates[EUR]
	if !ok {
		t.Fatal("expected EUR in response")
	}
	if eurRate <= 0 {
		t.Errorf("expected EUR rate to be > 0, got %v", eurRate)
	}

	sekRate, ok := data.Rates[SEK]
	if !ok {
		t.Fatal("expected SEK in response")
	}
	if sekRate <= 0 {
		t.Errorf("expected SEK rate to be > 0, got %v", sekRate)
	}
}

func TestGetSelectedExchangeRates_AllCurrencies(t *testing.T) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	currencyClient := NewCurrencyClient(httpClient)

	allCurrencies := []CurrencyCode{
		NOK, AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN,
		BAM, BBD, BDT, BGN, BHD, BIF, BMD, BND, BOB, BRL, BSD,
		BTN, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNH, CNY,
		COP, CRC, CUP, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ERN,
		ETB, EUR, FJD, FKP, FOK, GBP, GEL, GGP, GHS, GIP, GMD,
		GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, IMP,
		INR, IQD, IRR, ISK, JEP, JMD, JOD, JPY, KES, KGS, KHR,
		KID, KMF, KRW, KWD, KYD, KZT, LAK, LBP, LKR, LRD, LSL,
		LYD, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRU, MUR, MVR,
		MWK, MXN, MYR, MZN, NAD, NGN, NIO, NPR, NZD, OMR, PAB,
		PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF,
		SAR, SBD, SCR, SDG, SEK, SGD, SHP, SLE, SLL, SOS, SRD,
		SSP, STN, SYP, SZL, THB, TJS, TMT, TND, TOP, TRY, TTD,
		TVD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VES, VND, VUV,
		WST, XAF, XCD, XCG, XDR, XOF, XPF, YER, ZAR, ZMW, ZWG,
		ZWL,
	}

	req := Currency_InformationRequest{
		BaseCurrency: NOK,
		Currencies:   allCurrencies,
	}

	data, err := currencyClient.GetSelectedExchangeRates(req)
	if err != nil {
		t.Fatal(err)
	}

	if data.BaseCurrency != NOK {
		t.Fatalf("expected base currency to be NOK, got %q", data.BaseCurrency)
	}

	if len(data.Rates) != len(allCurrencies) {
		t.Fatalf("expected %d rates, got %d", len(allCurrencies), len(data.Rates))
	}

	for _, currency := range allCurrencies {
		rate, ok := data.Rates[currency]
		if !ok {
			t.Errorf("expected %s in response", currency)
			continue
		}

		if currency == NOK {
			if rate != 1 {
				t.Errorf("expected NOK rate to be 1, got %v", rate)
			}
			continue
		}

		if rate <= 0 {
			t.Errorf("expected %s rate to be > 0, got %v", currency, rate)
		}
	}
}
