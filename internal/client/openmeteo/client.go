package client

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WeatherClient interface {
	GetInfo(req Weather_InformationRequest) (Weather_INT_Response, error)
}

type weatherClient struct {
	httpClient *http.Client
	limiter    *rate.Limiter
}

func NewWeatherClient(httpClient *http.Client) WeatherClient {
	return &weatherClient{
		httpClient: httpClient,
		limiter:    rate.NewLimiter(rate.Every(1*time.Second), 1),
	}
}

/*
This struct contains the input fields required to query the weather API.
Lat and Lng are always required.
Temperature and Precipitation decide which values should be requested.
*/
type Weather_InformationRequest struct {
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	Temperature   bool    `json:"temperature"`
	Precipitation bool    `json:"precipitation"`
}

/*
This struct represents the external API response shape from Open-Meteo.
Only include the fields relevant for decoding.
*/
type Weather_EXT_Response struct {
	Hourly struct {
		Time          []string  `json:"time"`
		Temperature2M []float64 `json:"temperature_2m"`
		Precipitation []float64 `json:"precipitation"`
	} `json:"hourly"`
}

/*
This is the internal response returned by this client.
It should contain the resolved country name and mean values.
*/
type Weather_INT_Response struct {
	MeanTemperature   *float64
	MeanPrecipitation *float64
}

/*
Constants used only in this file.
*/
const (
	base_url = "https://api.open-meteo.com/v1/forecast"

	temperature   = "temperature_2m"
	precipitation = "precipitation"
	timezone      = "auto"
)

/*
This function is called externally.
It validates input, then calls functions that;
  - builds the URL
  - performs the HTTP request,
  - decodes the response
  - computes mean values based on response

after which it returns an internal response.
*/
func (c *weatherClient) GetInfo(req Weather_InformationRequest) (Weather_INT_Response, error) {

	if req.Lat < -90 || req.Lat > 90 {
		return Weather_INT_Response{}, fmt.Errorf("invalid latitude: must be between -90 and 90")
	}

	if req.Lng < -180 || req.Lng > 180 {
		return Weather_INT_Response{}, fmt.Errorf("invalid longitude: must be between -180 and 180")
	}

	if !req.Temperature && !req.Precipitation {
		return Weather_INT_Response{}, fmt.Errorf("a request for no weather information was made")
	}

	// 2. Build full API URL from request
	fullURL, err := buildURL(req)
	if err != nil {
		return Weather_INT_Response{}, err
	}

	body, err := c.httpRequestFunction(fullURL)
	if err != nil {
		return Weather_INT_Response{}, err
	}

	decoded, err := decodeResponse(body)
	if err != nil {
		return Weather_INT_Response{}, err
	}

	var meanTemp *float64
	if req.Temperature {
		value, err := calculateMean(decoded.Hourly.Temperature2M)
		if err != nil {
			return Weather_INT_Response{}, err
		}
		meanTemp = &value
	}

	var meanPrecip *float64
	if req.Precipitation {
		value, err := calculateMean(decoded.Hourly.Precipitation)
		if err != nil {
			return Weather_INT_Response{}, err
		}
		meanPrecip = &value
	}

	response := Weather_INT_Response{
		MeanTemperature:   meanTemp,
		MeanPrecipitation: meanPrecip,
	}

	if response.MeanTemperature != nil {
		fmt.Printf("meantemp: %v\n", *response.MeanTemperature)
	}
	if response.MeanPrecipitation != nil {
		fmt.Printf("meanprecip: %v\n", *response.MeanPrecipitation)
	}

	return response, nil
}

/*
This function constructs the request URL for the weather API.
It only includes the hourly parameters requested by the booleans.
*/
func buildURL(req Weather_InformationRequest) (string, error) {
	fields := make([]string, 0)

	if req.Temperature {
		fields = append(fields, temperature)
	}
	if req.Precipitation {
		fields = append(fields, precipitation)
	}

	query := url.Values{}
	query.Set("latitude", fmt.Sprintf("%f", req.Lat))
	query.Set("longitude", fmt.Sprintf("%f", req.Lng))
	query.Set("hourly", strings.Join(fields, ","))
	query.Set("timezone", timezone)

	fullURL := base_url + "?" + query.Encode()

	return fullURL, nil
}

/*
This function performs the outbound HTTP GET request.
It should use the injected httpClient.
*/
func (c *weatherClient) httpRequestFunction(fullURL string) ([]byte, error) {
	if err := c.limiter.Wait(context.Background()); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openmeteo error: status=%s body=%s", resp.Status, strings.TrimSpace(string(body)))
	}

	return body, nil
}

/*
This function unmarshals the raw weather API response.
*/
func decodeResponse(body []byte) (Weather_EXT_Response, error) {
	var response Weather_EXT_Response

	if err := json.Unmarshal(body, &response); err != nil {
		return Weather_EXT_Response{}, err
	}

	return response, nil
}

/*
This helper computes the mean of a float slice.
*/
func calculateMean(values []float64) (float64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("cannot calculate mean of empty slice")
	}

	var sum float64
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values)), nil
}

/*
We might use this if we want to cache the immediate weather info
This is all volatile information.
*/
func Initialize() {

	// 1. Optionally preload non-volatile weather-related config
	// 2. Optionally preload country lookup data if needed
	// 3. Store in cache
}
