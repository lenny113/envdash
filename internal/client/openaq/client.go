package client

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type OpenAQClient interface {
	GetInfo(req OpenAQ_InformationRequest) (OpenAQ_INT_Response, error)
}

type openAQClient struct {
	httpClient *http.Client
	apiKey     string
	limiter    *rate.Limiter
}

func NewOpenAQClient(httpClient *http.Client, apiKey string) OpenAQClient {
	return &openAQClient{
		httpClient: httpClient,
		apiKey:     strings.TrimSpace(apiKey),
		limiter:    rate.NewLimiter(rate.Every(time.Second), 1),
	}
}

/*
This struct contains the input fields required to query the OpenAQ API.
ISOCode is always required.
PM25 and PM10 decide which values should be requested.
*/
type OpenAQ_InformationRequest struct {
	ISOCode string `json:"isoCode"`
	PM25    bool   `json:"pm25"`
	PM10    bool   `json:"pm10"`
}

/*
This struct represents the external API response shape from OpenAQ.
Only include the fields relevant for decoding.
*/
type OpenAQ_EXT_Response struct {
	Meta struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
		Found int `json:"found"`
	} `json:"meta"`
	Results []struct {
		Value float64 `json:"value"`
	} `json:"results"`
}

/*
This is the internal response returned by this client.
It contains the requested mean values for the selected country.
*/
type OpenAQ_INT_Response struct {
	MeanPM25 *float64 `json:"meanPm25"`
	MeanPM10 *float64 `json:"meanPm10"`
}

/*
Constants used only in this file.
*/
const (
	base_url = "https://api.openaq.org/v3"

	pm10ParameterID = 1
	pm25ParameterID = 2
	maxLimit        = 1000

	latestQuery = "/parameters/%d/latest"
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
func (c *openAQClient) GetInfo(req OpenAQ_InformationRequest) (OpenAQ_INT_Response, error) {
	isoCode := strings.ToUpper(strings.TrimSpace(req.ISOCode))
	if isoCode == "" {
		return OpenAQ_INT_Response{}, fmt.Errorf("missing required iso code")
	}

	if !req.PM25 && !req.PM10 {
		return OpenAQ_INT_Response{}, fmt.Errorf("a request for no air quality information was made")
	}

	response := OpenAQ_INT_Response{}

	if req.PM25 {
		value, err := c.fetchMeanForParameter(isoCode, pm25ParameterID)
		if err != nil {
			return OpenAQ_INT_Response{}, err
		}
		response.MeanPM25 = &value
	}

	if req.PM10 {
		value, err := c.fetchMeanForParameter(isoCode, pm10ParameterID)
		if err != nil {
			return OpenAQ_INT_Response{}, err
		}
		response.MeanPM10 = &value
	}

	return response, nil
}

/*
This helper fetches the first page for a parameter and computes the mean value.
*/
func (c *openAQClient) fetchMeanForParameter(isoCode string, parameterID int) (float64, error) {
	fullURL, err := buildURL(isoCode, parameterID, 1)
	if err != nil {
		return 0, err
	}

	body, err := c.httpRequestFunction(fullURL)
	if err != nil {
		return 0, err
	}

	decoded, err := decodeResponse(body)
	if err != nil {
		return 0, err
	}

	values := make([]float64, 0, len(decoded.Results))
	for _, result := range decoded.Results {
		values = append(values, result.Value)
	}

	return calculateMean(values)
}

/*
This function constructs the request URL for the OpenAQ API.
It requests the latest values for a single parameter in a single country.
*/
func buildURL(isoCode string, parameterID int, page int) (string, error) {
	isoCode = strings.ToUpper(strings.TrimSpace(isoCode))
	if isoCode == "" {
		return "", fmt.Errorf("missing required iso code")
	}

	if page < 1 {
		return "", fmt.Errorf("invalid page number")
	}

	path := fmt.Sprintf(latestQuery, parameterID)

	query := url.Values{}
	query.Set("iso", isoCode)
	query.Set("limit", strconv.Itoa(maxLimit))
	query.Set("page", strconv.Itoa(page))

	fullURL := base_url + path + "?" + query.Encode()
	return fullURL, nil
}

/*
This function performs the outbound HTTP GET request.
It uses the injected httpClient and the API key from the environment.
*/
func (c *openAQClient) httpRequestFunction(fullURL string) ([]byte, error) {
	if err := c.limiter.Wait(context.Background()); err != nil {
		return nil, err
	}

	if strings.TrimSpace(c.apiKey) == "" {
		return nil, fmt.Errorf("missing OPENAQ API key")
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openaq error: status=%s body=%s", resp.Status, strings.TrimSpace(string(body)))
	}

	return body, nil
}

/*
This function unmarshals the raw OpenAQ API response.
*/
func decodeResponse(body []byte) (OpenAQ_EXT_Response, error) {
	var response OpenAQ_EXT_Response

	if err := json.Unmarshal(body, &response); err != nil {
		return OpenAQ_EXT_Response{}, err
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
We might use this if we want to cache air quality information.
This is volatile information.
*/
func Initialize() {

	// 1. Optionally preload frequently requested countries
	// 2. Optionally preload PM2.5 and PM10 means
	// 3. Store in cache
}
