package models

type RegisterWebhook struct {
	Url     string `json:"url"`
	Country string `json:"country"`
	Event   string `json:"event"`
}

type RegisteredWebhookResponse struct {
	Id      string `json:"id"`
	Country string `json:"country"`
	Event   string `json:"event"`
	Time    string `json:"time"`
}
