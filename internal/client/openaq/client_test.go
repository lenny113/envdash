package client

import "testing"

func TestBuildURL_Valid(t *testing.T) {
	fullURL, err := buildURL("no", pm25ParameterID, 1)
	if err != nil {
		t.Fatal(err)
	}

	expected := base_url + "/parameters/2/latest?iso=NO&limit=1000&page=1"
	if fullURL != expected {
		t.Fatalf("expected URL %q, got %q", expected, fullURL)
	}
}

func TestBuildURL_TrimsAndUppercasesISOCode(t *testing.T) {
	fullURL, err := buildURL(" no ", pm10ParameterID, 2)
	if err != nil {
		t.Fatal(err)
	}

	expected := base_url + "/parameters/1/latest?iso=NO&limit=1000&page=2"
	if fullURL != expected {
		t.Fatalf("expected URL %q, got %q", expected, fullURL)
	}
}

func TestBuildURL_MissingISOCode(t *testing.T) {
	_, err := buildURL("", pm25ParameterID, 1)
	if err == nil {
		t.Fatal("expected error when iso code is missing")
	}
}

func TestBuildURL_InvalidPage(t *testing.T) {
	_, err := buildURL("NO", pm25ParameterID, 0)
	if err == nil {
		t.Fatal("expected error when page is invalid")
	}
}

func TestDecodeResponse_Valid(t *testing.T) {
	body := []byte(`{
		"meta": {
			"page": 1,
			"limit": 1000,
			"found": 3
		},
		"results": [
			{"value": 10.5},
			{"value": 20.0},
			{"value": 30.5}
		]
	}`)

	data, err := decodeResponse(body)
	if err != nil {
		t.Fatal(err)
	}

	if data.Meta.Page != 1 {
		t.Fatalf("expected page 1, got %d", data.Meta.Page)
	}

	if data.Meta.Limit != 1000 {
		t.Fatalf("expected limit 1000, got %d", data.Meta.Limit)
	}

	if data.Meta.Found != 3 {
		t.Fatalf("expected found 3, got %d", data.Meta.Found)
	}

	if len(data.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(data.Results))
	}

	if data.Results[0].Value != 10.5 {
		t.Errorf("expected first value 10.5, got %v", data.Results[0].Value)
	}

	if data.Results[1].Value != 20.0 {
		t.Errorf("expected second value 20.0, got %v", data.Results[1].Value)
	}

	if data.Results[2].Value != 30.5 {
		t.Errorf("expected third value 30.5, got %v", data.Results[2].Value)
	}
}

func TestDecodeResponse_InvalidJSON(t *testing.T) {
	body := []byte(`{"meta":`)

	_, err := decodeResponse(body)
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestCalculateMean_Valid(t *testing.T) {
	values := []float64{10, 20, 30}

	mean, err := calculateMean(values)
	if err != nil {
		t.Fatal(err)
	}

	if mean != 20 {
		t.Fatalf("expected mean 20, got %v", mean)
	}
}

func TestCalculateMean_EmptySlice(t *testing.T) {
	_, err := calculateMean([]float64{})
	if err == nil {
		t.Fatal("expected error when calculating mean of empty slice")
	}
}

func TestInternalResponse_PM25Only(t *testing.T) {
	value := 12.34

	response := OpenAQ_INT_Response{
		MeanPM25: &value,
		MeanPM10: nil,
	}

	if response.MeanPM25 == nil {
		t.Fatal("expected MeanPM25 to be set")
	}

	if *response.MeanPM25 != 12.34 {
		t.Fatalf("expected MeanPM25 12.34, got %v", *response.MeanPM25)
	}

	if response.MeanPM10 != nil {
		t.Fatal("expected MeanPM10 to be nil")
	}
}

func TestInternalResponse_PM25AndPM10(t *testing.T) {
	pm25 := 11.1
	pm10 := 22.2

	response := OpenAQ_INT_Response{
		MeanPM25: &pm25,
		MeanPM10: &pm10,
	}

	if response.MeanPM25 == nil || response.MeanPM10 == nil {
		t.Fatal("expected both mean values to be set")
	}

	if *response.MeanPM25 != 11.1 {
		t.Fatalf("expected MeanPM25 11.1, got %v", *response.MeanPM25)
	}

	if *response.MeanPM10 != 22.2 {
		t.Fatalf("expected MeanPM10 22.2, got %v", *response.MeanPM10)
	}
}
