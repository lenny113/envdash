# Clients
Clients in this context refers to the files in the client directory. These are responsible for translating requests for information to http request to the respective apis.

Each client is separated into its own package and each client is responsible for handling its own api.
By splitting it like this it allows us to manage handlers, interfaces and ratelimiting per api, which helps us not abuse the external apis.

## Workflow.
Each client has an entrypoint which accepts a struct with boolean values dictating which information fields are requested. The exception to this is the currency api which accepts a base currency code.

After the request has been processed, they build a url that will fit the query. If possible we do not request more information than we need.

After this an http request for the relevant information is sent, this is then decoded and returned using the respective EXT_ response struct.

## Usages in the project

The APIs are designed to only be used by the cache struct in order to protect access to external apis. 
The exception to this is the status endpoint that monitors the connection to external apis, as this needs direct acess to ohld updated information.

## How to use the clients

Normal usage of the clients is through the cache's `requestFromCache` function, which checks whether the requested information is already cached before requesting it from the relevant APIS

For direct usage:

### Currency

The currency entry point is the `GetSelectedExchangeRates` function. It takes a `baseCurrency string` with the International Organization for Standardization (ISO) code for the currency.

It will give a response of:

```go
type Currency_INT_Response struct {
	BaseCurrency string             `json:"baseCurrency"`
	Rates        map[string]float64 `json:"rates"`
}
```

### RestCountries

The Rest Countries entry point is the `GetCountryInfo` function. It takes this request:

```go
type RestCountries_InformationRequest struct {
	BaseCountry string `json:"baseCountry"`
	ISOCode     string `json:"isoCode"`
	Name        bool   `json:"country"`
	CCA2        bool   `json:"cca2"`
	Capital     bool   `json:"capital"`
	Coordinates bool   `json:"coordinates"`
	Population  bool   `json:"population"`
	Area        bool   `json:"area"`
	Borders     bool   `json:"borders"`
	Currency    bool   `json:"currency"`
}
```

It will give a response of:

```go
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
```

### OpenAQ

The OpenAQ entry point is the `GetInfo` function. It takes this request:

```go
type OpenAQ_InformationRequest struct {
	ISOCode string `json:"isoCode"`
	PM25    bool   `json:"pm25"`
	PM10    bool   `json:"pm10"`
}
```

It will give a response of:

```go
type OpenAQ_INT_Response struct {
	MeanPM25 *float64 `json:"meanPm25"`
	MeanPM10 *float64 `json:"meanPm10"`
}
```

### Open-Meteo

The Open-Meteo entry point is the `GetInfo` function. It takes this request:

```go
type Weather_InformationRequest struct {
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	Temperature   bool    `json:"temperature"`
	Precipitation bool    `json:"precipitation"`
}
```

It will give a response of:

```go
type Weather_INT_Response struct {
	MeanTemperature   *float64
	MeanPrecipitation *float64
}
```

      
