package agify

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

type (
	// Client is the client to call agify.io
	Client struct {
		apiKey  string
		baseUrl string
		http    *http.Client
	}

	// clientDefaults is a struct used to hold the default values for the client
	clientDefaults struct {
		apiKey  string
		baseUrl string
		http    *http.Client
	}

	// ClientOption is a function that can be used to configure the client
	ClientOption func(*clientDefaults)

	// Prediction is the age prediction for a name
	Prediction struct {
		// Name is the name that was queried
		Name string `json:"name"`
		// Age is the predicted age
		Age int `json:"age"`
		// Count is the number of people with the same name
		Count int `json:"count"`
		// Country is the country that was queried
		Country string `json:"country_id"`
	}

	// RateLimit is the rate limiting information from the API
	RateLimit struct {
		Limit     string
		Remaining string
		Reset     string
	}

	// errorResponse is the error response from the agify API
	errorResponse struct {
		Error string `json:"error"`
	}
)

// WithApiKey overrides the default API key
func WithUrl(baseUrl string) ClientOption {
	return func(client *clientDefaults) {
		client.baseUrl = baseUrl
	}
}

// WithApiKey overrides the default API key
func WithApiKey(apiKey string) ClientOption {
	return func(client *clientDefaults) {
		client.apiKey = apiKey
	}
}

// WithClient overrides the default http client
func WithClient(httpClient *http.Client) ClientOption {
	return func(client *clientDefaults) {
		client.http = httpClient
	}
}

// NewClient creates a client to call agify.io
// By default, the client will use the public API URL without an API key.
// The default configuration can be overridden by passing in options.
func NewClient(opts ...ClientOption) *Client {
	// We use the default option to prevent Client options from having access to private data in the client
	defaults := &clientDefaults{
		apiKey:  "",
		baseUrl: "https://api.agify.io",
		http:    &http.Client{},
	}

	for _, opt := range opts {
		opt(defaults)
	}

	return &Client{
		apiKey:  defaults.apiKey,
		baseUrl: defaults.baseUrl,
		http:    defaults.http,
	}
}

// Predict returns the age probability for a name
func (client *Client) Predict(name string) (*Prediction, *RateLimit, error) {
	return client.PredictWithCountry(name, "")
}

// PredictWithCountry returns the age probability for a name in a country
func (client *Client) PredictWithCountry(name string, country string) (*Prediction, *RateLimit, error) {
	url, _ := url.Parse(client.baseUrl)
	values := url.Query()

	values.Add("name", name)

	if country != "" {
		values.Add("country_id", country)
	}

	if client.apiKey != "" {
		values.Add("apikey", client.apiKey)
	}

	url.RawQuery = values.Encode()

	body, rateLimit, err := client.get(url.String())

	if err != nil {
		return nil, rateLimit, err
	}

	var prediction Prediction
	err = json.Unmarshal(body, &prediction)

	if err != nil {
		return nil, rateLimit, err
	}

	return &prediction, rateLimit, nil
}

// BatchPredict returns the age probability for a list of names
func (client *Client) BatchPredict(names []string) (*[]Prediction, *RateLimit, error) {
	return client.BatchPredictWithCountry(names, "")
}

// BatchPredict returns the age probability for a list of names in a country
func (client *Client) BatchPredictWithCountry(names []string, country string) (*[]Prediction, *RateLimit, error) {
	url, _ := url.Parse(client.baseUrl)
	values := url.Query()

	values.Add("country_id", country)

	for _, name := range names {
		values.Add("name[]", name)
	}

	if client.apiKey != "" {
		values.Add("apikey", client.apiKey)
	}

	url.RawQuery = values.Encode()
	body, rateLimit, err := client.get(url.String())

	if err != nil {
		return nil, rateLimit, err
	}

	var predictions []Prediction
	err = json.Unmarshal(body, &predictions)

	if err != nil {
		return nil, rateLimit, err
	}

	return &predictions, rateLimit, nil
}

// get makes the API request and returns the response body
func (client *Client) get(url string) ([]byte, *RateLimit, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, nil, err
	}

	rateLimit := &RateLimit{
		Limit:     resp.Header.Get("X-Rate-Limit-Limit"),
		Remaining: resp.Header.Get("X-Rate-Limit-Remaining"),
		Reset:     resp.Header.Get("X-Rate-Reset"),
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var resp errorResponse
		err = json.Unmarshal(body, &resp)

		if err != nil {
			return nil, rateLimit, err
		}

		return nil, rateLimit, errors.New(resp.Error)
	}

	if err != nil {
		return nil, rateLimit, err
	}

	return body, rateLimit, nil
}
