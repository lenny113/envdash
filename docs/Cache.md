# Cache

The cache in this context refers to the cache implementation in the store package. It is responsible for coordinating access to data before external API clients are called. The cache is intended to be used as the single source of truth for application data access, so lookup logic, freshness rules, synchronization, and upstream fetch orchestration are kept in one place instead of being copied across handlers or services. This reduces repeated logic and helps performance by avoiding unnecessary external requests.  

Each cache entry is separated into fields, where every field is wrapped in a generic `Field[T]` structure. A field stores the value, whether the value is present, when it was last updated, and how stale it is allowed to become. This makes staleness detection happen per field instead of per entry, which means the cache can reuse fresh data while only refetching the parts that are missing or stale.  

The cache itself stores entries in memory and protects access with a `sync.RWMutex`. This is done so reads and writes can be coordinated safely in one place, instead of pushing mutex logic and state ownership out into the rest of the project.

## Workflow

The cache entry point is the `RequestFromCache` function. It first validates and normalizes the incoming request. This includes checking that at least one identifier is present, ensuring at least one field has been requested, and normalizing values such as CCA2 and currency codes.  

After the request has been normalized, the cache checks the in-memory entries for matching data. Freshness is evaluated with the `isFresh` helper, which only accepts fields that are present, have been updated before, and are still within their configured staleness window. Country-related data is treated as relatively stable, while weather, air quality, and currency data are given shorter staleness windows.  

If the cache does not already contain all the required fresh data, the intended workflow is to check Firestore next, and then fetch any remaining missing fields from the relevant upstream clients. After that, the cache re-checks the available data and returns a completed response. Firestore-backed caching is not implemented yet, but it is clearly planned in the infrastructure through `InitializeCache` and the documented request flow.  

## Usages in the project

The cache is designed to be the main access layer for application data. It owns the external clients for Rest Countries, Open-Meteo, currency, and OpenAQ, and is responsible for deciding whether data should come from memory, from future persistent storage, or from an upstream API call. This helps protect external APIs from being called directly in many places and keeps fetch policy centralized.  

## How to use the cache

Normal usage of the cache is through the `RequestFromCache` function.

It takes this request:

```go
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
```

It will give a response of:

```go
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
```



The cache also uses an internal field-based entry structure:

```go
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
```

## Features to be implemented

Firestore caching is not implemented yet, but it should be well supported by the existing infrastructure. `InitializeCache` is already intended as the startup location for loading persisted values, and the documented cache flow already includes a Firestore check before upstream fetches.  

In addition cache can be initialized by pulling long lived information at program startup. There is architecture for this, but it was not implemented due to time constraints.