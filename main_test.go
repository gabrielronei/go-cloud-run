package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Unit tests: temperature conversion ---

func TestCelsiusToFahrenheit(t *testing.T) {
	cases := []struct{ c, want float64 }{
		{0, 32},
		{100, 212},
		{-40, -40},
		{28.5, 83.3},
	}
	for _, tc := range cases {
		got := tc.c*1.8 + 32
		if got != tc.want {
			t.Errorf("C=%.1f: got %.1f, want %.1f", tc.c, got, tc.want)
		}
	}
}

func TestCelsiusToKelvin(t *testing.T) {
	cases := []struct{ c, want float64 }{
		{0, 273},
		{100, 373},
		{28.5, 301.5},
	}
	for _, tc := range cases {
		got := tc.c + 273
		if got != tc.want {
			t.Errorf("C=%.1f: got %.1f, want %.1f", tc.c, got, tc.want)
		}
	}
}

// --- Integration tests: HTTP handler with mock servers ---

func setupMockServers(t *testing.T, cepBody string, cepStatus int, weatherBody string, weatherStatus int) (*httptest.Server, *httptest.Server) {
	t.Helper()

	viaCEPServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(cepStatus)
		w.Write([]byte(cepBody))
	}))

	weatherServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(weatherStatus)
		w.Write([]byte(weatherBody))
	}))

	viaCEPBase = viaCEPServer.URL
	weatherBase = weatherServer.URL

	return viaCEPServer, weatherServer
}

func TestHandlerSuccess(t *testing.T) {
	viacep, weather := setupMockServers(t,
		`{"localidade":"São Paulo"}`, 200,
		`{"current":{"temp_c":25.0}}`, 200,
	)
	defer viacep.Close()
	defer weather.Close()

	req := httptest.NewRequest(http.MethodGet, "/01310100", nil)
	rec := httptest.NewRecorder()
	weatherHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var out WeatherOutput
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if out.TempC != 25.0 {
		t.Errorf("expected temp_C=25.0, got %.1f", out.TempC)
	}
	wantF := 25.0*1.8 + 32
	if out.TempF != wantF {
		t.Errorf("expected temp_F=%.1f, got %.1f", wantF, out.TempF)
	}
	wantK := 25.0 + 273
	if out.TempK != wantK {
		t.Errorf("expected temp_K=%.1f, got %.1f", wantK, out.TempK)
	}
}

func TestHandlerInvalidCEP(t *testing.T) {
	cases := []string{"/123", "/abcdefgh", "/12345678901", "/"}
	for _, path := range cases {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		weatherHandler(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("path %q: expected 422, got %d", path, rec.Code)
		}
		if rec.Body.String() != "invalid zipcode" {
			t.Errorf("path %q: expected 'invalid zipcode', got %q", path, rec.Body.String())
		}
	}
}

func TestHandlerCEPNotFound(t *testing.T) {
	viacep, weather := setupMockServers(t,
		`{"erro":"true"}`, 200,
		``, 200,
	)
	defer viacep.Close()
	defer weather.Close()

	req := httptest.NewRequest(http.MethodGet, "/00000000", nil)
	rec := httptest.NewRecorder()
	weatherHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if rec.Body.String() != "can not find zipcode" {
		t.Errorf("expected 'can not find zipcode', got %q", rec.Body.String())
	}
}
