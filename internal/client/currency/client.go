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

type CurrencyClient interface {
	GetSelectedExchangeRates(baseCurrency string) (Currency_INT_Response, error)
}

type currencyClient struct {
	httpClient *http.Client
	limiter    *rate.Limiter
}

func NewCurrencyClient(httpClient *http.Client) CurrencyClient {
	return &currencyClient{
		httpClient: httpClient,
		limiter:    rate.NewLimiter(rate.Every(1*time.Second), 1),
	}
}

/*
This struct represents the external API response shape from the currency API.
Only include the fields relevant for decoding.
*/
type Currency_EXT_Response struct {
	Result             string             `json:"result"`
	Provider           string             `json:"provider"`
	Documentation      string             `json:"documentation"`
	TermsOfUse         string             `json:"terms_of_use"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
	TimeLastUpdateUTC  string             `json:"time_last_update_utc"`
	TimeNextUpdateUnix int64              `json:"time_next_update_unix"`
	TimeNextUpdateUTC  string             `json:"time_next_update_utc"`
	TimeEOLUnix        int64              `json:"time_eol_unix"`
	BaseCode           string             `json:"base_code"`
	Rates              map[string]float64 `json:"rates"`
}

/*
This is the internal response returned by this client.
It contains the base currency and all exchange rates returned by the API.
*/
type Currency_INT_Response struct {
	BaseCurrency string             `json:"baseCurrency"`
	Rates        map[string]float64 `json:"rates"`
}

/*
Constants used only in this file.
*/
const (
	base_url = "http://129.241.150.113:9090/currency/"
)

/*
This function is called externally.
It validates input, then calls functions that;
  - builds the URL
  - performs the HTTP request
  - decodes the response

after which it returns an internal response.
It also prints the conversions for debugging.
*/
func (c *currencyClient) GetSelectedExchangeRates(baseCurrency string) (Currency_INT_Response, error) {
	if strings.TrimSpace(baseCurrency) == "" {
		return Currency_INT_Response{}, fmt.Errorf("missing required base currency")
	}

	fullURL, err := buildURL(baseCurrency)
	if err != nil {
		return Currency_INT_Response{}, err
	}

	body, err := c.httpRequestFunction(fullURL)
	if err != nil {
		return Currency_INT_Response{}, err
	}

	decoded, err := decodeResponse(body)
	if err != nil {
		return Currency_INT_Response{}, err
	}

	response := Currency_INT_Response{
		BaseCurrency: strings.ToUpper(strings.TrimSpace(decoded.BaseCode)),
		Rates:        decoded.Rates,
	}

	return response, nil
}

/*
This function constructs the request URL for the currency API.
It requests all rates for the provided base currency.
*/
func buildURL(baseCurrency string) (string, error) {
	baseCurrency = strings.ToUpper(strings.TrimSpace(baseCurrency))
	if baseCurrency == "" {
		return "", fmt.Errorf("missing required base currency")
	}

	fullURL := base_url + url.PathEscape(baseCurrency)
	return fullURL, nil
}

/*
This function performs the outbound HTTP GET request.
It uses the injected httpClient.
*/
func (c *currencyClient) httpRequestFunction(fullURL string) ([]byte, error) {
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
		return nil, fmt.Errorf("currency api error: status=%s body=%s", resp.Status, strings.TrimSpace(string(body)))
	}

	return body, nil
}

/*
This function unmarshals the raw currency API response.
*/
func decodeResponse(body []byte) (Currency_EXT_Response, error) {
	var response Currency_EXT_Response

	if err := json.Unmarshal(body, &response); err != nil {
		return Currency_EXT_Response{}, err
	}

	if strings.TrimSpace(response.BaseCode) == "" {
		return Currency_EXT_Response{}, fmt.Errorf("currency api response missing base currency")
	}

	if response.Rates == nil {
		return Currency_EXT_Response{}, fmt.Errorf("currency api response missing rates")
	}

	if response.Result != "" && response.Result != "success" {
		return Currency_EXT_Response{}, fmt.Errorf("currency api returned result %q", response.Result)
	}

	return response, nil
}
