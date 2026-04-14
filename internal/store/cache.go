package store

import (
	currencyclient "assignment-2/internal/client/currency"
	aqclient "assignment-2/internal/client/openaq"
	weatherclient "assignment-2/internal/client/openmeteo"
	countryclient "assignment-2/internal/client/restcountries"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Field[T any] struct {
	Value       T
	Present     bool
	LastUpdated time.Time
	Staleness   time.Duration
}

type Entry struct {
	CountryName Field[string]
	CCA2        Field[string]
	Capital     Field[string]
	Coordinates Field[[]float64]
	Population  Field[int64]
	Area        Field[float64]
	Borders     Field[[]string]

	MeanTemperature   Field[float64]
	MeanPrecipitation Field[float64]

	CurrencyBase  Field[string]
	CurrencyRates map[string]*Field[float64]

	MeanPM25 Field[float64]
	MeanPM10 Field[float64]
}

// Request recieved to the cache
type CacheExternalRequest struct {
	Name string
	CCA2 string

	CountryName       bool
	CountryCCA2       bool
	Capital           bool
	Coordinates       bool
	Population        bool
	Area              bool
	Borders           bool
	MeanTemperature   bool
	MeanPrecipitation bool
	MeanPM25          bool
	MeanPM10          bool
	CurrencyBase      bool
	CurrencyRates     []string
}

// Response going out from the cache
type CacheResponse struct {
	CountryName       *string
	CountryCCA2       *string
	Capital           *string
	Coordinates       *[]float64
	Population        *int64
	Area              *float64
	Borders           *[]string
	MeanTemperature   *float64
	MeanPrecipitation *float64
	MeanPM25          *float64
	MeanPM10          *float64
	CurrencyBase      *string
	CurrencyRates     map[string]float64
}

type Cache struct {
	mu sync.RWMutex

	Entries []*Entry

	countryClient  countryclient.RestCountriesClient
	weatherClient  weatherclient.WeatherClient
	currencyClient currencyclient.CurrencyClient
	aqClient       aqclient.OpenAQClient
}

/*
NewCache constructs an empty in-memory cache and stores the external clients
used later when missing fields must be fetched from upstream services.
*/
func NewCache(
	countryClient countryclient.RestCountriesClient,
	weatherClient weatherclient.WeatherClient,
	currencyClient currencyclient.CurrencyClient,
	aqClient aqclient.OpenAQClient,
) *Cache {
	return &Cache{
		Entries:        []*Entry{},
		countryClient:  countryClient,
		weatherClient:  weatherClient,
		currencyClient: currencyClient,
		aqClient:       aqClient,
	}
}

/*
InitializeCache is the intended startup constructor.

Right now it only creates a fresh cache instance.
Later this is the place to preload persisted values from Firestore and any
other non-volatile data you want ready before handling requests.
*/
func InitializeCache(
	countryClient countryclient.RestCountriesClient,
	weatherClient weatherclient.WeatherClient,
	currencyClient currencyclient.CurrencyClient,
	aqClient aqclient.OpenAQClient,
) *Cache {
	cache := NewCache(countryClient, weatherClient, currencyClient, aqClient)

	// TODO:
	// 1. Fetch persisted values from Firestore
	// 2. Preload non-volatile fields if wanted

	return cache
}

/*
newEntry creates one empty cache entry with field names and default staleness
values already assigned.

Country metadata is treated as relatively stable.
Weather is short-lived.
Currency is medium-lived.
*/
func newEntry() *Entry {
	return &Entry{
		CountryName: Field[string]{
			Staleness: 24 * time.Hour,
		},
		CCA2: Field[string]{
			Staleness: 24 * time.Hour,
		},
		Capital: Field[string]{
			Staleness: 24 * time.Hour,
		},
		Coordinates: Field[[]float64]{
			Staleness: 24 * time.Hour,
		},
		Population: Field[int64]{
			Staleness: 24 * time.Hour,
		},
		Area: Field[float64]{
			Staleness: 24 * time.Hour,
		},
		Borders: Field[[]string]{
			Staleness: 24 * time.Hour,
		},
		MeanTemperature: Field[float64]{
			Staleness: 30 * time.Minute,
		},
		MeanPrecipitation: Field[float64]{
			Staleness: 30 * time.Minute,
		},
		CurrencyBase: Field[string]{
			Staleness: 1 * time.Hour,
		},
		MeanPM25: Field[float64]{
			Staleness: 30 * time.Minute,
		},
		MeanPM10: Field[float64]{
			Staleness: 30 * time.Minute,
		},
		CurrencyRates: make(map[string]*Field[float64]),
	}
}

/*
isFresh returns true only when a field:
- has a value
- has been updated at least once
- and is not older than its allowed staleness window

used when checking data, meaning we do not update unless we need to.
We considered updating whenever the staleness reaches 0, this would marginally increase the robustness and response time.
however doing so would create unnecessary noise for the third party APIs.
*/
func isFresh[T any](field Field[T]) bool {
	if !field.Present {
		return false
	}
	if field.LastUpdated.IsZero() {
		return false
	}
	if field.Staleness <= 0 {
		return true
	}
	return time.Since(field.LastUpdated) < field.Staleness
}

/*
normalizeRequest validates the external request and normalizes the identifier
and currency codes.

Country Code Alpha-2 (CCA2) is uppercased.
Currency codes are uppercased and deduplicated.
The request must contain at least one identifier and at least one requested field.
*/
func normalizeRequest(req CacheExternalRequest) (CacheExternalRequest, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.CCA2 = strings.ToUpper(strings.TrimSpace(req.CCA2))

	if req.Name == "" && req.CCA2 == "" {
		return req, fmt.Errorf("name or CCA2 is required")
	}

	seen := make(map[string]struct{})
	normalizedRates := make([]string, 0, len(req.CurrencyRates))

	for _, code := range req.CurrencyRates {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			continue
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		normalizedRates = append(normalizedRates, code)
	}
	req.CurrencyRates = normalizedRates

	if requestEmpty(req) {
		return req, fmt.Errorf("at least one requested field is required")
	}

	return req, nil
}

/*
requestEmpty reports whether a request asks for no data fields at all.

Identifiers alone are not enough; at least one data field must be requested.
*/
func requestEmpty(req CacheExternalRequest) bool {
	return !req.CountryName &&
		!req.CountryCCA2 &&
		!req.Capital &&
		!req.Coordinates &&
		!req.Population &&
		!req.Area &&
		!req.Borders &&
		!req.MeanTemperature &&
		!req.MeanPrecipitation &&
		!req.CurrencyBase &&
		!req.MeanPM25 &&
		!req.MeanPM10 &&
		len(req.CurrencyRates) == 0
}

/*
findEntry searches the in-memory cache for an existing entry.

It prefers a Country Code Alpha-2 (CCA2) match when one is supplied, but can
also match on country name when needed.
*/
func (c *Cache) findEntry(name, cca2 string) *Entry {
	name = strings.TrimSpace(name)
	cca2 = strings.ToUpper(strings.TrimSpace(cca2))

	for _, entry := range c.Entries {
		if entry == nil {
			continue
		}

		if cca2 != "" && entry.CCA2.Present && strings.EqualFold(entry.CCA2.Value, cca2) {
			return entry
		}

		if name != "" && entry.CountryName.Present && strings.EqualFold(entry.CountryName.Value, name) {
			return entry
		}
	}

	return nil
}

/*
entryStoredLocked checks whether a specific entry pointer is already present
in the cache slice.

The caller must already hold the cache lock when calling this helper.
*/
func (c *Cache) entryStoredLocked(target *Entry) bool {
	for _, entry := range c.Entries {
		if entry == target {
			return true
		}
	}
	return false
}

/*
copyAvailableRequestedFields copies only the fresh fields that are already
available from src into out.

For every field that is successfully copied, the matching request flag is
cleared so later stages know it no longer needs to be fetched.
*/
func copyAvailableRequestedFields(src *Entry, req *CacheExternalRequest, out *Entry) {
	if src == nil {
		return
	}

	if req.CountryName && isFresh(src.CountryName) {
		out.CountryName = src.CountryName
		req.CountryName = false
	}

	if req.CountryCCA2 && isFresh(src.CCA2) {
		out.CCA2 = src.CCA2
		req.CountryCCA2 = false
	}

	if req.Capital && isFresh(src.Capital) {
		out.Capital = src.Capital
		req.Capital = false
	}

	if req.Coordinates && isFresh(src.Coordinates) {
		out.Coordinates = src.Coordinates
		out.Coordinates.Value = append([]float64(nil), src.Coordinates.Value...)
		req.Coordinates = false
	}

	if req.Population && isFresh(src.Population) {
		out.Population = src.Population
		req.Population = false
	}

	if req.Area && isFresh(src.Area) {
		out.Area = src.Area
		req.Area = false
	}

	if req.Borders && isFresh(src.Borders) {
		out.Borders = src.Borders
		out.Borders.Value = append([]string(nil), src.Borders.Value...)
		req.Borders = false
	}

	if req.MeanTemperature && isFresh(src.MeanTemperature) {
		out.MeanTemperature = src.MeanTemperature
		req.MeanTemperature = false
	}

	if req.MeanPrecipitation && isFresh(src.MeanPrecipitation) {
		out.MeanPrecipitation = src.MeanPrecipitation
		req.MeanPrecipitation = false
	}

	if req.CurrencyBase && isFresh(src.CurrencyBase) {
		out.CurrencyBase = src.CurrencyBase
		req.CurrencyBase = false
	}
	if req.MeanPM25 && isFresh(src.MeanPM25) {
		out.MeanPM25 = src.MeanPM25
		req.MeanPM25 = false
	}

	if req.MeanPM10 && isFresh(src.MeanPM10) {
		out.MeanPM10 = src.MeanPM10
		req.MeanPM10 = false
	}

	if len(req.CurrencyRates) > 0 && src.CurrencyRates != nil {
		remainingRates := make([]string, 0, len(req.CurrencyRates))

		if out.CurrencyRates == nil {
			out.CurrencyRates = make(map[string]*Field[float64])
		}

		for _, code := range req.CurrencyRates {
			field, ok := src.CurrencyRates[code]
			if !ok || field == nil || !isFresh(*field) {
				remainingRates = append(remainingRates, code)
				continue
			}

			copied := *field
			out.CurrencyRates[code] = &copied
		}

		req.CurrencyRates = remainingRates
	}
}

/*
checkCache performs the in-memory lookup stage.

It returns:
- a partial response containing any fresh fields already in memory
- a reduced request containing only still-missing fields
- the matching backing entry, if one exists
*/
func (c *Cache) checkCache(req CacheExternalRequest) (*Entry, CacheExternalRequest, *Entry) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	response := newEntry()
	remaining := req
	entry := c.findEntry(req.Name, req.CCA2)

	copyAvailableRequestedFields(entry, &remaining, response)

	return response, remaining, entry
}

/*
checkFirestore is the persistence lookup stage.

It is intentionally left as a stub for now, but the intended behavior is the
same as checkCache: return what exists, and leave only missing fields pending.
*/
func (c *Cache) checkFirestore(req CacheExternalRequest) (*Entry, CacheExternalRequest, error) {
	// TODO: implement Firestore lookup later
	return newEntry(), req, nil
}

/*
needsCountryFetch reports whether a request requires data from the
REST Countries client.

A country fetch is needed when:
- country fields are requested directly
- weather is requested and coordinates are missing or stale
- currency data is requested and CurrencyBase is missing or stale
*/
func needsCountryFetch(req CacheExternalRequest, entry *Entry) bool {
	if req.CountryName || req.CountryCCA2 || req.Capital || req.Coordinates || req.Population || req.Area || req.Borders {
		return true
	}

	if (req.MeanTemperature || req.MeanPrecipitation) &&
		(entry == nil || !isFresh(entry.Coordinates) || len(entry.Coordinates.Value) < 2) {
		return true
	}

	if req.CurrencyBase || len(req.CurrencyRates) > 0 {
		if entry == nil || !isFresh(entry.CurrencyBase) {
			return true
		}
	}

	if req.MeanPM25 || req.MeanPM10 {
		if entry == nil || !isFresh(entry.CCA2) {
			return true
		}
	}

	return false
}

/*
buildCountryRequest translates the external cache request into the specific
REST Countries request needed for missing data.

It always requests country identity fields so the entry can remain
consistently identified by name and Country Code Alpha-2 (CCA2).
*/
func buildCountryRequest(req CacheExternalRequest, entry *Entry) (countryclient.RestCountries_InformationRequest, error) {
	out := countryclient.RestCountries_InformationRequest{}

	switch {
	case strings.TrimSpace(req.CCA2) != "":
		out.ISOCode = strings.ToUpper(strings.TrimSpace(req.CCA2))
	case entry != nil && entry.CCA2.Present:
		out.ISOCode = strings.ToUpper(strings.TrimSpace(entry.CCA2.Value))
	case strings.TrimSpace(req.Name) != "":
		out.BaseCountry = strings.TrimSpace(req.Name)
	case entry != nil && entry.CountryName.Present:
		out.BaseCountry = strings.TrimSpace(entry.CountryName.Value)
	default:
		return out, fmt.Errorf("missing identifier for country request")
	}

	// Always fetch identity fields when calling country service.
	out.Name = true
	out.CCA2 = true

	if req.Capital {
		out.Capital = true
	}
	if req.Coordinates || req.MeanTemperature || req.MeanPrecipitation {
		out.Coordinates = true
	}
	if req.Population {
		out.Population = true
	}
	if req.Area {
		out.Area = true
	}
	if req.Borders {
		out.Borders = true
	}
	if req.CurrencyBase || len(req.CurrencyRates) > 0 {
		out.Currency = true
	}

	return out, nil
}

func applyAQData(entry *Entry, resp aqclient.OpenAQ_INT_Response) {
	now := time.Now()

	if resp.MeanPM25 != nil {
		entry.MeanPM25.Value = *resp.MeanPM25
		entry.MeanPM25.Present = true
		entry.MeanPM25.LastUpdated = now
	}

	if resp.MeanPM10 != nil {
		entry.MeanPM10.Value = *resp.MeanPM10
		entry.MeanPM10.Present = true
		entry.MeanPM10.LastUpdated = now
	}
}

/*
applyCountryData writes the country client response into the cache entry and
updates timestamps for every field that was successfully received.
*/
func applyCountryData(entry *Entry, resp countryclient.RestCountries_INT_Response) {
	now := time.Now()

	if resp.Country != nil {
		entry.CountryName.Value = *resp.Country
		entry.CountryName.Present = true
		entry.CountryName.LastUpdated = now
	}

	if resp.CCA2 != nil {
		entry.CCA2.Value = strings.ToUpper(strings.TrimSpace(*resp.CCA2))
		entry.CCA2.Present = true
		entry.CCA2.LastUpdated = now
	}

	if resp.Capital != nil {
		entry.Capital.Value = *resp.Capital
		entry.Capital.Present = true
		entry.Capital.LastUpdated = now
	}

	if resp.Coordinates != nil {
		entry.Coordinates.Value = append([]float64(nil), (*resp.Coordinates)...)
		entry.Coordinates.Present = true
		entry.Coordinates.LastUpdated = now
	}

	if resp.Population != nil {
		entry.Population.Value = *resp.Population
		entry.Population.Present = true
		entry.Population.LastUpdated = now
	}

	if resp.Area != nil {
		entry.Area.Value = *resp.Area
		entry.Area.Present = true
		entry.Area.LastUpdated = now
	}

	if resp.Borders != nil {
		entry.Borders.Value = append([]string(nil), (*resp.Borders)...)
		entry.Borders.Present = true
		entry.Borders.LastUpdated = now
	}

	// this one is more complex in order to handle multiple currencies returned.
	if resp.Currencies != nil && len(*resp.Currencies) > 0 {
		codes := append([]string(nil), (*resp.Currencies)...)
		sort.Strings(codes)

		entry.CurrencyBase.Value = codes[0]
		entry.CurrencyBase.Present = true
		entry.CurrencyBase.LastUpdated = now
	}
}

/*
applyWeatherData writes the weather client response into the cache entry and
refreshes timestamps for the weather fields received.
*/
func applyWeatherData(entry *Entry, resp weatherclient.Weather_INT_Response) {
	now := time.Now()

	if resp.MeanTemperature != nil {
		entry.MeanTemperature.Value = *resp.MeanTemperature
		entry.MeanTemperature.Present = true
		entry.MeanTemperature.LastUpdated = now
	}

	if resp.MeanPrecipitation != nil {
		entry.MeanPrecipitation.Value = *resp.MeanPrecipitation
		entry.MeanPrecipitation.Present = true
		entry.MeanPrecipitation.LastUpdated = now
	}
}

/*
applyCurrencyData writes the requested currency data into the cache entry.

The currency client returns many rates, but only the explicitly requested
codes are stored in CurrencyRates here.
*/
func applyCurrencyData(entry *Entry, base string, resp currencyclient.Currency_INT_Response, requestedRates []string) {
	now := time.Now()

	entry.CurrencyBase.Value = strings.ToUpper(strings.TrimSpace(base))
	entry.CurrencyBase.Present = true
	entry.CurrencyBase.LastUpdated = now

	if entry.CurrencyRates == nil {
		entry.CurrencyRates = make(map[string]*Field[float64])
	}

	for _, code := range requestedRates {
		rate, ok := resp.Rates[code]
		if !ok {
			continue
		}

		entry.CurrencyRates[code] = &Field[float64]{
			Value:       rate,
			Present:     true,
			LastUpdated: now,
			Staleness:   1 * time.Hour,
		}
	}
}

/*
sendGet performs all upstream fetching for whatever fields are still missing.

Order:
1. Fetch country data if needed
2. Fetch weather if requested
3. Fetch air quality if requested
3. Fetch currency if requested

Every successful upstream response is written back into the same cache entry.
*/
func (c *Cache) sendGet(req CacheExternalRequest, entry *Entry) error {
	if entry == nil {
		entry = newEntry()
	}

	if needsCountryFetch(req, entry) {
		if c.countryClient == nil {
			return fmt.Errorf("country client is not configured")
		}

		countryReq, err := buildCountryRequest(req, entry)
		if err != nil {
			return err
		}

		resp, err := c.countryClient.GetCountryInfo(countryReq)
		if err != nil {
			return err
		}

		c.mu.Lock()
		applyCountryData(entry, resp)
		if !c.entryStoredLocked(entry) {
			c.Entries = append(c.Entries, entry)
		}
		c.mu.Unlock()
	} else {
		c.mu.Lock()
		if !c.entryStoredLocked(entry) {
			c.Entries = append(c.Entries, entry)
		}
		c.mu.Unlock()
	}

	if req.MeanTemperature || req.MeanPrecipitation {
		if c.weatherClient == nil {
			return fmt.Errorf("weather client is not configured")
		}
		if !entry.Coordinates.Present || len(entry.Coordinates.Value) < 2 {
			return fmt.Errorf("coordinates missing for weather request")
		}

		weatherReq := weatherclient.Weather_InformationRequest{
			Lat:           entry.Coordinates.Value[0],
			Lng:           entry.Coordinates.Value[1],
			Temperature:   req.MeanTemperature,
			Precipitation: req.MeanPrecipitation,
		}

		resp, err := c.weatherClient.GetInfo(weatherReq)
		if err != nil {
			return err
		}

		c.mu.Lock()
		applyWeatherData(entry, resp)
		c.mu.Unlock()
	}

	if req.CurrencyBase || len(req.CurrencyRates) > 0 {
		if c.currencyClient == nil {
			return fmt.Errorf("currency client is not configured")
		}

		base := ""
		if !entry.CurrencyBase.Present || !isFresh(entry.CurrencyBase) {
			return fmt.Errorf("fresh currency base missing for country")
		}
		base = entry.CurrencyBase.Value

		resp, err := c.currencyClient.GetSelectedExchangeRates(base)
		if err != nil {
			return err
		}

		c.mu.Lock()
		applyCurrencyData(entry, base, resp, req.CurrencyRates)
		c.mu.Unlock()
	}

	if req.MeanPM25 || req.MeanPM10 {
		if c.aqClient == nil {
			return fmt.Errorf("air quality client is not configured")
		}

		if !entry.CCA2.Present || !isFresh(entry.CCA2) {
			return fmt.Errorf("fresh CCA2 missing for air quality request")
		}

		aqReq := aqclient.OpenAQ_InformationRequest{
			ISOCode: entry.CCA2.Value,
			PM25:    req.MeanPM25,
			PM10:    req.MeanPM10,
		}

		resp, err := c.aqClient.GetInfo(aqReq)
		if err != nil {
			return err
		}

		c.mu.Lock()
		applyAQData(entry, resp)
		c.mu.Unlock()
	}

	return nil
}

/*
applyResponseData copies fields from one entry into another response entry.

This is used after the Firestore stage so that any fields recovered there are
combined into the response already being built.
*/
func applyResponseData(dst, src *Entry) {
	if src == nil {
		return
	}

	if src.CountryName.Present {
		dst.CountryName = src.CountryName
	}
	if src.CCA2.Present {
		dst.CCA2 = src.CCA2
	}
	if src.Capital.Present {
		dst.Capital = src.Capital
	}
	if src.Coordinates.Present {
		dst.Coordinates = src.Coordinates
		dst.Coordinates.Value = append([]float64(nil), src.Coordinates.Value...)
	}
	if src.Population.Present {
		dst.Population = src.Population
	}
	if src.Area.Present {
		dst.Area = src.Area
	}
	if src.Borders.Present {
		dst.Borders = src.Borders
		dst.Borders.Value = append([]string(nil), src.Borders.Value...)
	}
	if src.MeanTemperature.Present {
		dst.MeanTemperature = src.MeanTemperature
	}
	if src.MeanPrecipitation.Present {
		dst.MeanPrecipitation = src.MeanPrecipitation
	}
	if src.CurrencyBase.Present {
		dst.CurrencyBase = src.CurrencyBase
	}
	if len(src.CurrencyRates) > 0 {
		if dst.CurrencyRates == nil {
			dst.CurrencyRates = make(map[string]*Field[float64])
		}
		for code, field := range src.CurrencyRates {
			if field == nil {
				continue
			}
			copied := *field
			dst.CurrencyRates[code] = &copied
		}
	}

	if src.MeanPM25.Present {
		dst.MeanPM25 = src.MeanPM25
	}
	if src.MeanPM10.Present {
		dst.MeanPM10 = src.MeanPM10
	}
}

func entryToCacheResponse(entry *Entry) *CacheResponse {
	if entry == nil {
		return &CacheResponse{}
	}

	resp := &CacheResponse{}

	if entry.CountryName.Present {
		value := entry.CountryName.Value
		resp.CountryName = &value
	}

	if entry.CCA2.Present {
		value := entry.CCA2.Value
		resp.CountryCCA2 = &value
	}

	if entry.Capital.Present {
		value := entry.Capital.Value
		resp.Capital = &value
	}

	if entry.Coordinates.Present {
		value := append([]float64(nil), entry.Coordinates.Value...)
		resp.Coordinates = &value
	}

	if entry.Population.Present {
		value := entry.Population.Value
		resp.Population = &value
	}

	if entry.Area.Present {
		value := entry.Area.Value
		resp.Area = &value
	}

	if entry.Borders.Present {
		value := append([]string(nil), entry.Borders.Value...)
		resp.Borders = &value
	}

	if entry.MeanTemperature.Present {
		value := entry.MeanTemperature.Value
		resp.MeanTemperature = &value
	}

	if entry.MeanPrecipitation.Present {
		value := entry.MeanPrecipitation.Value
		resp.MeanPrecipitation = &value
	}

	if entry.CurrencyBase.Present {
		value := entry.CurrencyBase.Value
		resp.CurrencyBase = &value
	}

	if len(entry.CurrencyRates) > 0 {
		resp.CurrencyRates = make(map[string]float64, len(entry.CurrencyRates))
		for code, field := range entry.CurrencyRates {
			if field == nil || !field.Present {
				continue
			}
			resp.CurrencyRates[code] = field.Value
		}
	}

	if entry.MeanPM25.Present {
		value := entry.MeanPM25.Value
		resp.MeanPM25 = &value
	}

	if entry.MeanPM10.Present {
		value := entry.MeanPM10.Value
		resp.MeanPM10 = &value
	}

	return resp
}

/*
RequestFromCache is the main public entrypoint.

Flow:
1. Validate and normalize the external request
2. Check in-memory cache
3. Check Firestore
4. Fetch any still-missing fields from upstream services
5. Re-check cache and return the completed response
*/
func (c *Cache) RequestFromCache(req CacheExternalRequest) (*CacheResponse, error) {
	if c == nil {
		return nil, fmt.Errorf("cache is nil")
	}

	req, err := normalizeRequest(req)
	if err != nil {
		return nil, err
	}

	response, remaining, entry := c.checkCache(req)
	if requestEmpty(remaining) {
		return entryToCacheResponse(response), nil
	}

	firestoreResp, remaining, err := c.checkFirestore(remaining)
	if err != nil {
		return nil, err
	}
	applyResponseData(response, firestoreResp)

	if requestEmpty(remaining) {
		return entryToCacheResponse(response), nil
	}

	if err := c.sendGet(remaining, entry); err != nil {
		return nil, err
	}

	finalResp, finalRemaining, _ := c.checkCache(req)
	if !requestEmpty(finalRemaining) {
		return entryToCacheResponse(finalResp), fmt.Errorf("some requested fields are still missing after fetch")
	}

	return entryToCacheResponse(finalResp), nil
}
