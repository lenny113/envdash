package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

type RestCountriesClient interface {
	GetCountryInfo(req RestCountries_InformationRequest) (RestCountries_INT_Response, error)
}
type restCountriesClient struct {
	httpClient *http.Client
	limiter    *rate.Limiter
}

func NewRestCountriesClient(httpClient *http.Client) RestCountriesClient {
	return &restCountriesClient{
		httpClient: httpClient,
		limiter:    rate.NewLimiter(rate.Every(1*time.Second), 1),
	}
}

/*
this struct contains the relevant fields that we can request from RESTCOUNTRIES API.
It should contain the fields that are relevant when building the API.
*/
type RestCountries_InformationRequest struct {
	BaseCountry string `json:"baseCountry"`
	ISOCode     string `json:"isoCode"`
	Name        bool   `json:"country"` // common name
	CCA2        bool   `json:"cca2"`
	Capital     bool   `json:"capital"`
	Coordinates bool   `json:"coordinates"`
	Population  bool   `json:"population"`
	Area        bool   `json:"area"`
	Borders     bool   `json:"borders"`
	Currency    bool   `json:"currency"`
}

// the json we expect to recieve from the restcountries API request
// this struct contains the response from an iso request
type RestCountries_EXT_ISO struct {
	Name struct {
		Common string `json:"common"`
	} `json:"name"`

	CCA2       string    `json:"cca2"`
	Capital    []string  `json:"capital"`
	LatLng     []float64 `json:"latlng"`
	Population int64     `json:"population"`
	Area       float64   `json:"area"`
	Borders    []string  `json:"borders"`
	Currencies map[string]struct {
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
	} `json:"currencies"`
}

// when reqeusting using name, we recieve a slice instead of a json object, so this variable is needed to handle
// unmarshaling this.
type RestCountries_EXT_Name []RestCountries_EXT_ISO

// We return this struct from the file, this should be written to cache.
type RestCountries_INT_Response struct {
	Country     *string
	CCA2        *string
	Capital     *string
	Coordinates *[]float64
	Population  *int64
	Area        *float64
	Borders     *[]string
	Currencies  *[]string
}

/*
Constants used in the context of the REST countries API
These are only used in this file so they are non exportable by intent.
*/
const (
	base_url = "http://129.241.150.113:8080/v3.1"

	cca2        = "cca2"
	name        = "name"
	capital     = "capital"
	coordinates = "latlng"
	population  = "population"
	area        = "area"
	borders     = "borders"
	currencies  = "currencies"

	iso_query  = "/alpha/"
	name_query = "/name/"
)

/*
This function is called externally, returning the requested info from the restcountries API base URL.
It assumes that appropriate input validation has already happened.
*/

func (c *restCountriesClient) GetCountryInfo(req RestCountries_InformationRequest) (RestCountries_INT_Response, error) {

	fullURL, askedUsingIso, err := buildURL(req)
	if err != nil {
		return RestCountries_INT_Response{}, err
	}

	// TODO: see if we can decouple the request here somehow.
	// TODO: add more robust error handling, we should inspect the error and return an appropriate http request.
	// Right now we only forward the error returned from a get request.
	if err := c.limiter.Wait(context.Background()); err != nil {
		return RestCountries_INT_Response{}, err
	}

	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return RestCountries_INT_Response{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return RestCountries_INT_Response{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return RestCountries_INT_Response{}, fmt.Errorf("restcountries error: status=%s body=%s", resp.Status, strings.TrimSpace(string(body)))
	}

	data, err := decodeRESTCountriesResponse(body, askedUsingIso)
	if err != nil {
		return RestCountries_INT_Response{}, err
	}

	return data, nil
}

func buildURL(req RestCountries_InformationRequest) (string, bool, error) {
	var path string
	askedUsingIso := true

	if strings.TrimSpace(req.ISOCode) != "" {
		path = iso_query + url.PathEscape(strings.TrimSpace(req.ISOCode))
	} else if strings.TrimSpace(req.BaseCountry) != "" {
		path = name_query + url.PathEscape(strings.TrimSpace(req.BaseCountry))
		askedUsingIso = false
	} else {
		// TODO: add logger here
		return "", false, fmt.Errorf("missing required identifier: isoCode or baseCountry")
	}

	if !req.CCA2 && !req.Name && !req.Capital && !req.Coordinates && !req.Population && !req.Area && !req.Borders && !req.Currency {
		// TODO: add logger here
		return "", false, fmt.Errorf("a request for no information was made")
	}

	fields := make([]string, 0)

	if req.Name {
		fields = append(fields, name)
	}
	if req.CCA2 {
		fields = append(fields, cca2)
	}
	if req.Capital {
		fields = append(fields, capital)
	}
	if req.Coordinates {
		fields = append(fields, coordinates)
	}
	if req.Population {
		fields = append(fields, population)
	}
	if req.Area {
		fields = append(fields, area)
	}
	if req.Borders {
		fields = append(fields, borders)
	}
	if req.Currency {
		fields = append(fields, currencies)
	}

	path += "?fields=" + strings.Join(fields, ",")
	fullURL := base_url + path

	return fullURL, askedUsingIso, nil
}

func decodeRESTCountriesResponse(body []byte, askedUsingISO bool) (RestCountries_INT_Response, error) {

	src, err := resolveResponseType(body, askedUsingISO)
	if err != nil {
		return RestCountries_INT_Response{}, err
	}

	result := RestCountries_INT_Response{}

	if src.Name.Common != "" {
		country := src.Name.Common
		result.Country = &country
	}
	if src.CCA2 != "" {
		cca2 := src.CCA2
		result.CCA2 = &cca2
	}

	if len(src.Capital) > 0 {
		capital := src.Capital[0]
		result.Capital = &capital
	}

	if len(src.LatLng) >= 2 {
		coords := append([]float64(nil), src.LatLng...)
		result.Coordinates = &coords
	}

	population := src.Population
	result.Population = &population

	area := src.Area
	result.Area = &area
	if len(src.Currencies) > 0 {
		codes := make([]string, 0, len(src.Currencies))
		for code := range src.Currencies {
			codes = append(codes, code)
		}
		result.Currencies = &codes
	}
	return result, nil
}

// Returns a single struct based on the requested information from REST countries
// RESTcountries returns a different response based on whether we ask using an iso code or a country name.
// I really dont like this solution, but i could not find a better one.
// TODO: see if we can find a cleaner implementation.
func resolveResponseType(body []byte, askedUsingISO bool) (RestCountries_EXT_ISO, error) {
	if askedUsingISO {
		var response RestCountries_EXT_ISO
		if err := json.Unmarshal(body, &response); err != nil {
			return RestCountries_EXT_ISO{}, err
		}
		return response, nil
	}

	var response RestCountries_EXT_Name
	if err := json.Unmarshal(body, &response); err != nil {
		return RestCountries_EXT_ISO{}, err
	}
	if len(response) == 0 {
		// TODO: add logger here
		return RestCountries_EXT_ISO{}, fmt.Errorf("empty response from restcountries")
	}

	return response[0], nil
}

/*
	 This function should be called at server startup that will request non volatile information
	 and cache this in the program

	 non-volatile info is:
		BaseCountry
		ISOCode
		Capital
		Coordinates
		Area
		Borders
*/
func Initialize() {

}
