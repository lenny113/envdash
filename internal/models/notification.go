package models

type RegisterWebhook struct {
	Url                   string                 `json:"url"`
	Country               string                 `json:"country"`
	Event                 string                 `json:"event"`
	ThresholdNotification *ThresholdNotification `json:"threshold,omitempty" firestore:"threshold,omitempty"`
	User                  string                 `json:"-" firestore:"user"`
	Time                  string                 `json:"-" firestore:"time"`
}

type ThresholdNotification struct {
	Field    string  `json:"field"`
	Operator string  `json:"operator"`
	Value    float64 `json:"value"`
}

type RegisteredWebhookResponse struct {
	Id string `json:"id"`
}

type AllRegisteredWebhook struct {
	Id string `json:"id"`
	RegisterWebhook
}

type ThresholdDetails struct {
	Field          string  `json:"field"`
	Operator       string  `json:"operator"`
	ThresholdValue float64 `json:"threshold"`
	MeasuredValue  float64 `json:"measuredValue"`
}
