package store

import "testing"

func TestRequestFromCache_FullFlow(t *testing.T) {
	cache := newTestCache(t)

	req := CacheExternalRequest{
		CCA2: "NO",

		CountryName:       true,
		CountryCCA2:       true,
		Capital:           true,
		Coordinates:       true,
		Population:        true,
		Area:              true,
		Borders:           true,
		MeanTemperature:   true,
		MeanPrecipitation: true,
		CurrencyBase:      true,
		MeanPM25:          true,
		MeanPM10:          true,
		CurrencyRates:     []string{"EUR", "SEK"},
	}

	resp, err := cache.RequestFromCache(req)
	if err != nil {
		t.Fatalf("RequestFromCache returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if resp.CountryName == nil || *resp.CountryName != "Norway" {
		t.Fatalf("expected CountryName Norway, got %#v", resp.CountryName)
	}
	if resp.CountryCCA2 == nil || *resp.CountryCCA2 != "NO" {
		t.Fatalf("expected CountryCCA2 NO, got %#v", resp.CountryCCA2)
	}
	if resp.Capital == nil || *resp.Capital != "Oslo" {
		t.Fatalf("expected Capital Oslo, got %#v", resp.Capital)
	}
	if resp.Coordinates == nil || len(*resp.Coordinates) != 2 {
		t.Fatalf("expected 2 coordinates, got %#v", resp.Coordinates)
	}
	if (*resp.Coordinates)[0] != 62 {
		t.Fatalf("expected latitude 62, got %v", (*resp.Coordinates)[0])
	}
	if (*resp.Coordinates)[1] != 10 {
		t.Fatalf("expected longitude 10, got %v", (*resp.Coordinates)[1])
	}
	if resp.Population == nil || *resp.Population != 5379475 {
		t.Fatalf("expected Population 5379475, got %#v", resp.Population)
	}
	if resp.Area == nil || *resp.Area != 323802 {
		t.Fatalf("expected Area 323802, got %#v", resp.Area)
	}
	if resp.Borders == nil || len(*resp.Borders) != 3 {
		t.Fatalf("expected 3 borders, got %#v", resp.Borders)
	}
	if resp.MeanTemperature == nil {
		t.Fatal("expected MeanTemperature")
	}
	if resp.MeanPrecipitation == nil {
		t.Fatal("expected MeanPrecipitation")
	}
	if resp.CurrencyBase == nil || *resp.CurrencyBase != "NOK" {
		t.Fatalf("expected CurrencyBase NOK, got %#v", resp.CurrencyBase)
	}
	if resp.CurrencyRates == nil {
		t.Fatal("expected CurrencyRates map")
	}
	if _, ok := resp.CurrencyRates["EUR"]; !ok {
		t.Fatalf("expected EUR rate in CurrencyRates, got %#v", resp.CurrencyRates)
	}
	if _, ok := resp.CurrencyRates["SEK"]; !ok {
		t.Fatalf("expected SEK rate in CurrencyRates, got %#v", resp.CurrencyRates)
	}
	if resp.MeanPM25 == nil {
		t.Fatal("expected MeanPM25")
	}
	if *resp.MeanPM25 < 0 {
		t.Fatalf("expected MeanPM25 to be >= 0, got %v", *resp.MeanPM25)
	}

	if resp.MeanPM10 == nil {
		t.Fatal("expected MeanPM10")
	}
	if *resp.MeanPM10 < 0 {
		t.Fatalf("expected MeanPM10 to be >= 0, got %v", *resp.MeanPM10)
	}

}
