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
