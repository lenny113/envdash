package client
type restCountriesClient struct {
	httpClient *http.Client
}

func NewRestCountriesClient(httpClient *http.Client) RestCountriesClient {
	return &restCountriesClient{
		httpClient: httpClient,
	}
}
