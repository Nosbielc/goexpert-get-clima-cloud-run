package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
)

func TestIsValidCEP(t *testing.T) {
	tests := []struct {
		cep      string
		expected bool
	}{
		{"01310-100", true},
		{"01310100", true},
		{"123", false},
		{"1234567890", false},
		{"abcdefgh", false},
		{"", false},
	}

	for _, test := range tests {
		result := isValidCEP(test.cep)
		if result != test.expected {
			t.Errorf("isValidCEP(%s) = %v; expected %v", test.cep, result, test.expected)
		}
	}
}

func TestCelsiusToFahrenheit(t *testing.T) {
	tests := []struct {
		celsius  float64
		expected float64
	}{
		{0, 32},
		{100, 212},
		{25, 77},
		{-10, 14},
	}

	for _, test := range tests {
		result := celsiusToFahrenheit(test.celsius)
		if result != test.expected {
			t.Errorf("celsiusToFahrenheit(%f) = %f; expected %f", test.celsius, result, test.expected)
		}
	}
}

func TestCelsiusToKelvin(t *testing.T) {
	tests := []struct {
		celsius  float64
		expected float64
	}{
		{0, 273},
		{100, 373},
		{25, 298},
		{-273, 0},
	}

	for _, test := range tests {
		result := celsiusToKelvin(test.celsius)
		if result != test.expected {
			t.Errorf("celsiusToKelvin(%f) = %f; expected %f", test.celsius, result, test.expected)
		}
	}
}

func TestWeatherHandlerInvalidCEP(t *testing.T) {
	req, err := http.NewRequest("GET", "/weather/123", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/weather/{cep}", weatherHandler)

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}

	var response ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if response.Message != "invalid zipcode" {
		t.Errorf("handler returned unexpected message: got %v want %v", response.Message, "invalid zipcode")
	}
}

func TestWeatherHandlerNotFoundCEP(t *testing.T) {
	req, err := http.NewRequest("GET", "/weather/99999999", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/weather/{cep}", weatherHandler)

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	var response ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if response.Message != "can not find zipcode" {
		t.Errorf("handler returned unexpected message: got %v want %v", response.Message, "can not find zipcode")
	}
}

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// Test integration with valid CEP (requires WEATHER_API_KEY)
func TestWeatherHandlerValidCEP(t *testing.T) {
	// Skip if no API key is set
	if os.Getenv("WEATHER_API_KEY") == "" {
		t.Skip("WEATHER_API_KEY not set, skipping integration test")
	}

	req, err := http.NewRequest("GET", "/weather/01310-100", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/weather/{cep}", weatherHandler)

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response TemperatureResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	// Verify temperature conversion
	expectedF := celsiusToFahrenheit(response.TempC)
	expectedK := celsiusToKelvin(response.TempC)

	if response.TempF != expectedF {
		t.Errorf("Fahrenheit conversion incorrect: got %v want %v", response.TempF, expectedF)
	}

	if response.TempK != expectedK {
		t.Errorf("Kelvin conversion incorrect: got %v want %v", response.TempK, expectedK)
	}
}
