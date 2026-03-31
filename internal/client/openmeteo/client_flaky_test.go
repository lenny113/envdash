//go:build flaky
// +build flaky

package client

import (
	utils "assignment-2/internal/utils"
	"testing"
)

func TestGetInfo(t *testing.T) {
	httpClient := utils.NewHttpClient()
	weatherClient := NewWeatherClient(httpClient)

	req := Weather_InformationRequest{
		Lat:           59.91,
		Lng:           10.75,
		Temperature:   true,
		Precipitation: true,
	}

	data, err := weatherClient.GetInfo(req)
	if err != nil {
		t.Fatal(err)
	}

	if data.MeanTemperature == nil {
		t.Fatal("expected MeanTemperature in response")
	}

	if data.MeanPrecipitation == nil {
		t.Fatal("expected MeanPrecipitation in response")
	}

	if *data.MeanTemperature < -100 || *data.MeanTemperature > 100 {
		t.Errorf("expected MeanTemperature to be within a reasonable range, got %v", *data.MeanTemperature)
	}

	if *data.MeanPrecipitation < 0 {
		t.Errorf("expected MeanPrecipitation to be non-negative, got %v", *data.MeanPrecipitation)
	}
}

func TestGetInfo_TemperatureOnly(t *testing.T) {
	httpClient := utils.NewHttpClient()
	weatherClient := NewWeatherClient(httpClient)

	req := Weather_InformationRequest{
		Lat:           59.91,
		Lng:           10.75,
		Temperature:   true,
		Precipitation: false,
	}

	data, err := weatherClient.GetInfo(req)
	if err != nil {
		t.Fatal(err)
	}

	if data.MeanTemperature == nil {
		t.Fatal("expected MeanTemperature in response")
	}

	if data.MeanPrecipitation != nil {
		t.Fatal("expected MeanPrecipitation to be nil")
	}
}

func TestGetInfo_PrecipitationOnly(t *testing.T) {
	httpClient := utils.NewHttpClient()
	weatherClient := NewWeatherClient(httpClient)

	req := Weather_InformationRequest{
		Lat:           59.91,
		Lng:           10.75,
		Temperature:   false,
		Precipitation: true,
	}

	data, err := weatherClient.GetInfo(req)
	if err != nil {
		t.Fatal(err)
	}

	if data.MeanPrecipitation == nil {
		t.Fatal("expected MeanPrecipitation in response")
	}

	if data.MeanTemperature != nil {
		t.Fatal("expected MeanTemperature to be nil")
	}

	if *data.MeanPrecipitation < 0 {
		t.Errorf("expected MeanPrecipitation to be non-negative, got %v", *data.MeanPrecipitation)
	}
}
