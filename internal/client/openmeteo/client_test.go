package client

import (
	utils "assignment-2/internal/utils"
	"testing"
)

func TestGetInfo_InvalidLatitude(t *testing.T) {
	httpClient := utils.NewHttpClient()
	weatherClient := NewWeatherClient(httpClient)

	req := Weather_InformationRequest{
		Lat:           100,
		Lng:           10.75,
		Temperature:   true,
		Precipitation: true,
	}

	_, err := weatherClient.GetInfo(req)
	if err == nil {
		t.Fatal("expected error for invalid latitude")
	}
}

func TestGetInfo_InvalidLongitude(t *testing.T) {
	httpClient := utils.NewHttpClient()
	weatherClient := NewWeatherClient(httpClient)

	req := Weather_InformationRequest{
		Lat:           59.91,
		Lng:           200,
		Temperature:   true,
		Precipitation: true,
	}

	_, err := weatherClient.GetInfo(req)
	if err == nil {
		t.Fatal("expected error for invalid longitude")
	}
}

func TestGetInfo_NoRequestedFields(t *testing.T) {
	httpClient := utils.NewHttpClient()
	weatherClient := NewWeatherClient(httpClient)

	req := Weather_InformationRequest{
		Lat:           59.91,
		Lng:           10.75,
		Temperature:   false,
		Precipitation: false,
	}

	_, err := weatherClient.GetInfo(req)
	if err == nil {
		t.Fatal("expected error when no weather information was requested")
	}
}
