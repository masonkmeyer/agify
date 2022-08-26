package agify

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldCreateNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
}

func TestShouldGetPredictionForName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Rate-Limit-Limit", "1000")
		w.Header().Set("X-Rate-Limit-Remaining", "728")
		w.Header().Set("X-Rate-Reset", "15281")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"michael","age":70,"count":875,"country_id":"US"}`))
	}))
	defer server.Close()

	client := NewClient(WithUrl(server.URL))

	result, rateLimit, err := client.Predict("michael")
	assert.Nil(t, err)
	assert.Equal(t, 70, result.Age)
	assert.Equal(t, 875, result.Count)
	assert.Equal(t, "michael", result.Name)
	assert.Equal(t, "US", result.Country)

	assert.Equal(t, "1000", rateLimit.Limit)
	assert.Equal(t, "728", rateLimit.Remaining)
	assert.Equal(t, "15281", rateLimit.Reset)
}

func TestShouldGetErrorWhenUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{ "error": "Invalid API key" }`))
	}))
	defer server.Close()

	client := NewClient(WithUrl(server.URL))
	result, rateLimit, err := client.Predict("michael")

	assert.Nil(t, result)
	assert.NotNil(t, rateLimit)
	assert.NotNil(t, err)
}

func TestShouldGetErrorWhenTooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{ "error": "Request limit reached" }`))
	}))
	defer server.Close()

	client := NewClient(WithUrl(server.URL))
	result, rateLimit, err := client.Predict("michael")

	assert.Nil(t, result)
	assert.NotNil(t, rateLimit)
	assert.NotNil(t, err)
}

func TestShouldGetErrorWhenUnprocessable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{ "error": "Too many requests" }`))
	}))
	defer server.Close()

	client := NewClient(WithUrl(server.URL))
	result, rateLimit, err := client.Predict("michael")

	assert.Nil(t, result)
	assert.NotNil(t, rateLimit)
	assert.NotNil(t, err)
}

func TestShouldOverrideDefaults(t *testing.T) {
	client := NewClient(WithUrl("http://localhost:8080"), WithClient(&http.Client{}), WithApiKey("test-key"))
	assert.NotNil(t, client)
}
