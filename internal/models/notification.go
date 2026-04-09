package models

type RegisterWebhook struct {
	Url     string `json:"url"`
	Country string `json:"country"`
	Event   string `json:"event"`
}
