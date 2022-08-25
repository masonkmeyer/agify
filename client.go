package agify

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

type RateLimit struct {
	Limit     string
	Remaining string
	Reset     string
}

type Prediction struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Count   int    `json:"count"`
	Country string `json:"country_id"`
}

type Client struct {
	apiKey  *string
	baseUrl string
	http    *http.Client
}

type errorResponse struct {
	Error string `json:"error"`
}

// NewClient creates a new client for agify.io
func NewClient(apiKey *string, baseUrl string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseUrl: baseUrl,
		http:    &http.Client{},
	}
}

// Predict returns the age probabilty for a name
func (client *Client) Predict(name string) (*Prediction, *RateLimit, error) {
	return client.PredictWithCountry(name, "")
}

// PredictWithCountry returns the age probabilty for a name in a country
func (client *Client) PredictWithCountry(name string, country string) (*Prediction, *RateLimit, error) {
	url, _ := url.Parse(client.baseUrl)
	values := url.Query()

	values.Add("name", name)

	if country != "" {
		values.Add("country_id", country)
	}

	if client.apiKey != nil {
		values.Add("apikey", *client.apiKey)
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

// BatchPredict returns the age probabilty for a list of names
func (client *Client) BatchPredict(names []string) (*[]Prediction, *RateLimit, error) {
	return client.BatchPredictWithCountry(names, "")
}

// BatchPredict returns the age probabilty for a list of names in a country
func (client *Client) BatchPredictWithCountry(names []string, country string) (*[]Prediction, *RateLimit, error) {
	url, _ := url.Parse(client.baseUrl)
	values := url.Query()

	values.Add("country_id", country)

	for _, name := range names {
		values.Add("name[]", name)
	}

	if client.apiKey != nil {
		values.Add("apikey", *client.apiKey)
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
