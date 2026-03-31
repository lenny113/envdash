package models

type Authentication struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	ApiKey    string `json:"apiKey"`
	CreatedAt string `json:"createdAt"`
}
