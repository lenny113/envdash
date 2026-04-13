package client

import (
	"strings"
	"testing"
)

func TestBuildURL_WithISOCode(t *testing.T) {
	req := RestCountries_InformationRequest{
		ISOCode:     "NO",
		Name:        true,
		CCA2:        true,
		Capital:     true,
		Coordinates: true,
		Population:  true,
		Area:        true,
		Currency:    true,
		Borders:     true,
	}

	fullURL, askedUsingISO, err := buildURL(req)
	if err != nil {
		t.Fatal(err)
	}

	if !askedUsingISO {
		t.Fatal("expected askedUsingISO to be true")
	}

	expectedPrefix := base_url + "/alpha/NO?fields="
	if !strings.HasPrefix(fullURL, expectedPrefix) {
		t.Fatalf("expected URL to start with %q, got %q", expectedPrefix, fullURL)
	}

	expectedFields := []string{
		"name",
		"cca2",
		"capital",
		"latlng",
		"population",
		"area",
		"borders",
		"currencies",
	}

	for _, field := range expectedFields {
		if !strings.Contains(fullURL, field) {
			t.Errorf("expected URL to contain field %q, got %q", field, fullURL)
		}
	}
}

func TestBuildURL_WithBaseCountry(t *testing.T) {
	req := RestCountries_InformationRequest{
		BaseCountry: "Norway",
		Name:        true,
		CCA2:        true,
	}

	fullURL, askedUsingISO, err := buildURL(req)
	if err != nil {
		t.Fatal(err)
	}

	if askedUsingISO {
		t.Fatal("expected askedUsingISO to be false")
	}

	expectedPrefix := base_url + "/name/Norway?fields="
	if !strings.HasPrefix(fullURL, expectedPrefix) {
		t.Fatalf("expected URL to start with %q, got %q", expectedPrefix, fullURL)
	}

	if !strings.Contains(fullURL, "name") {
		t.Errorf("expected URL to contain field %q, got %q", "name", fullURL)
	}
	if !strings.Contains(fullURL, "cca2") {
		t.Errorf("expected URL to contain field %q, got %q", "cca2", fullURL)
	}
}

func TestBuildURL_MissingIdentifier(t *testing.T) {
	req := RestCountries_InformationRequest{
		Name: true,
	}

	_, _, err := buildURL(req)
	if err == nil {
		t.Fatal("expected error when neither ISOCode nor BaseCountry is set")
	}
}

func TestBuildURL_NoRequestedFields(t *testing.T) {
	req := RestCountries_InformationRequest{
		ISOCode: "NO",
	}

	_, _, err := buildURL(req)
	if err == nil {
		t.Fatal("expected error when no fields are requested")
	}
}

func TestDecodeRESTCountriesResponse_ISO(t *testing.T) {
	body := []byte(`{
	"name": { "common": "Norway" },
	"cca2": "NO",
	"capital": ["Oslo"],
	"latlng": [62, 10],
	"population": 5379475,
	"area": 323802,
	"borders": ["FIN", "SWE", "RUS"],
	"currencies": {
		"NOK": {
			"name": "Norwegian krone",
			"symbol": "kr"
		}
	}
}`)

	data, err := decodeRESTCountriesResponse(body, true)
	if err != nil {
		t.Fatal(err)
	}

	if data.Country == nil || *data.Country != "Norway" {
		t.Fatalf("expected Country to be Norway, got %#v", data.Country)
	}
	if data.CCA2 == nil || *data.CCA2 != "NO" {
		t.Fatalf("expected CCA2 to be NO, got %#v", data.CCA2)
	}
	if data.Capital == nil || *data.Capital != "Oslo" {
		t.Fatalf("expected Capital to be Oslo, got %#v", data.Capital)
	}
	if data.Coordinates == nil || len(*data.Coordinates) != 2 {
		t.Fatalf("expected 2 coordinates, got %#v", data.Coordinates)
	}
	if (*data.Coordinates)[0] != 62 {
		t.Errorf("expected latitude 62, got %v", (*data.Coordinates)[0])
	}
	if (*data.Coordinates)[1] != 10 {
		t.Errorf("expected longitude 10, got %v", (*data.Coordinates)[1])
	}
	if data.Population == nil || *data.Population != 5379475 {
		t.Fatalf("expected Population 5379475, got %#v", data.Population)
	}
	if data.Area == nil || *data.Area != 323802 {
		t.Fatalf("expected Area 323802, got %#v", data.Area)
	}
	if data.Borders == nil || len(*data.Borders) != 3 {
		t.Fatalf("expected 3 borders, got %#v", data.Borders)
	}
	if data.Currencies == nil || len(*data.Currencies) != 1 {
		t.Fatalf("expected 1 currency, got %#v", data.Currencies)
	}
	if (*data.Currencies)[0] != "NOK" {
		t.Fatalf("expected currency NOK, got %#v", *data.Currencies)
	}
}

func TestDecodeRESTCountriesResponse_Name(t *testing.T) {
	body := []byte(`[
	{
		"name": { "common": "Norway" },
		"cca2": "NO",
		"capital": ["Oslo"],
		"latlng": [62, 10],
		"population": 5379475,
		"area": 323802,
		"borders": ["FIN", "SWE", "RUS"],
		"currencies": {
			"NOK": {
				"name": "Norwegian krone",
				"symbol": "kr"
			}
		}
	}
]`)

	data, err := decodeRESTCountriesResponse(body, false)
	if err != nil {
		t.Fatal(err)
	}

	if data.Country == nil || *data.Country != "Norway" {
		t.Fatalf("expected Country to be Norway, got %#v", data.Country)
	}
	if data.CCA2 == nil || *data.CCA2 != "NO" {
		t.Fatalf("expected CCA2 to be NO, got %#v", data.CCA2)
	}
	if data.Currencies == nil || len(*data.Currencies) != 1 {
		t.Fatalf("expected 1 currency, got %#v", data.Currencies)
	}
	if (*data.Currencies)[0] != "NOK" {
		t.Fatalf("expected currency NOK, got %#v", *data.Currencies)
	}
}

func TestResolveResponseType_EmptyNameResponse(t *testing.T) {
	body := []byte(`[]`)

	_, err := resolveResponseType(body, false)
	if err == nil {
		t.Fatal("expected error for empty response")
	}
}
