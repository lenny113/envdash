//go:build flaky
// +build flaky

package client

import (
	utils "assignment-2/internal/utils"
	"testing"
)

func TestGetCountryInfo(t *testing.T) {
	httpClient := utils.NewHttpClient()
	restCountriesClient := NewRestCountriesClient(httpClient)

	req := RestCountries_InformationRequest{
		ISOCode:     "NO",
		CCA2:        true,
		Name:        true,
		Capital:     true,
		Coordinates: true,
		Population:  true,
		Area:        true,
		Borders:     true,
		Currency:    true,
	}

	data, err := restCountriesClient.GetCountryInfo(req)
	if err != nil {
		t.Fatal(err)
	}

	if data.CCA2 == nil {
		t.Fatal("expected CCA2 in response")
	}
	if *data.CCA2 != "NO" {
		t.Errorf("expected CCA2 to be NO, got %q", *data.CCA2)
	}

	if data.Country == nil {
		t.Fatal("expected Country in response")
	}
	if *data.Country != "Norway" {
		t.Errorf("expected Country to be Norway, got %q", *data.Country)
	}

	if data.Capital == nil {
		t.Fatal("expected Capital in response")
	}
	if *data.Capital != "Oslo" {
		t.Errorf("expected Capital to be Oslo, got %q", *data.Capital)
	}

	if data.Coordinates == nil {
		t.Fatal("expected Coordinates in response")
	}
	if len(*data.Coordinates) < 2 {
		t.Fatalf("expected at least 2 coordinates, got %v", *data.Coordinates)
	}
	if (*data.Coordinates)[0] != 62 {
		t.Errorf("expected latitude to be 62, got %v", (*data.Coordinates)[0])
	}
	if (*data.Coordinates)[1] != 10 {
		t.Errorf("expected longitude to be 10, got %v", (*data.Coordinates)[1])
	}

	if data.Population == nil {
		t.Fatal("expected Population in response")
	}

	if data.Area == nil {
		t.Fatal("expected Area in response")
	}
	if *data.Area != 323802 {
		t.Errorf("expected Area to be 323802, got %v", *data.Area)
	}

	if data.Borders == nil {
		t.Fatal("expected Borders in response")
	}
	expectedBorders := []string{"FIN", "SWE", "RUS"}
	if len(*data.Borders) != len(expectedBorders) {
		t.Fatalf("expected %d borders, got %d: %v", len(expectedBorders), len(*data.Borders), *data.Borders)
	}
	for i, expected := range expectedBorders {
		if (*data.Borders)[i] != expected {
			t.Errorf("expected border[%d] to be %q, got %q", i, expected, (*data.Borders)[i])
		}
	}

	if data.Currencies == nil {
		t.Fatal("expected Currencies in response")
	}

	foundNOK := false
	for _, code := range *data.Currencies {
		if code == "NOK" {
			foundNOK = true
			break
		}
	}
	if !foundNOK {
		t.Errorf("expected Currencies to contain NOK, got %v", *data.Currencies)
	}
}
