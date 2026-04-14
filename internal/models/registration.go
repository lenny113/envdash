package models

// Definition of a registration
type Registration struct {
	ID         string   `json:"id,omitempty" firestore:"id,omitempty"`
	Country    string   `json:"country" firestore:"country"`
	IsoCode    string   `json:"isoCode" firestore:"isoCode"`
	Features   Features `json:"features" firestore:"features"`
	LastChange string   `json:"lastChange" firestore:"lastChange"`
}

// Defining features that are part of a registration
type Features struct {
	Temperature      bool     `json:"temperature" firestore:"temperature"`
	Precipitation    bool     `json:"precipitation" firestore:"precipitation"`
	AirQuality       bool     `json:"airQuality" firestore:"airQuality"`
	Capital          bool     `json:"capital" firestore:"capital"`
	Coordinates      bool     `json:"coordinates" firestore:"coordinates"`
	Population       bool     `json:"population" firestore:"population"`
	Area             bool     `json:"area" firestore:"area"`
	TargetCurrencies []string `json:"targetCurrencies" firestore:"targetCurrencies"`
}

//Defintion of response when asking external api for countryname and isocode that will be used for validation

type CountryNameAndISO struct {
	CCA2 string `json:"cca2"`
	Name struct {
		Common   string `json:"common"`
		Official string `json:"official"`
	} `json:"name"`
}

// patch from user, where any field can be empty or have a value
type RegistrationPatch struct {
	Country  *string `json:"country,omitempty"`
	IsoCode  *string `json:"isoCode,omitempty"`
	Features *struct {
		Temperature      *bool     `json:"temperature,omitempty"`
		Precipitation    *bool     `json:"precipitation,omitempty"`
		AirQuality       *bool     `json:"airQuality,omitempty"`
		Capital          *bool     `json:"capital,omitempty"`
		Coordinates      *bool     `json:"coordinates,omitempty"`
		Population       *bool     `json:"population,omitempty"`
		Area             *bool     `json:"area,omitempty"`
		TargetCurrencies *[]string `json:"targetCurrencies,omitempty"`
	} `json:"features,omitempty"`
}

// Struct to store the currencyAPI response for valdiation
type CurrencyAPIResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}
