package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

var (
	httpClient  = &http.Client{Timeout: 10 * time.Second}
	cepRegex    = regexp.MustCompile(`^\d{8}$`)
	viaCEPBase  = "https://viacep.com.br/ws"
	weatherBase = "https://api.weatherapi.com/v1"
)

type WeatherOutput struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type viaCEPResponse struct {
	Localidade string `json:"localidade"`
	Erro       string `json:"erro,omitempty"`
}

type weatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func getCityByCEP(cep string) (string, error) {
	resp, err := httpClient.Get(fmt.Sprintf("%s/%s/json/", viaCEPBase, cep))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return "", fmt.Errorf("cep not found")
	}

	var data viaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if data.Erro == "true" || data.Localidade == "" {
		return "", fmt.Errorf("cep not found")
	}

	return data.Localidade, nil
}

func getTemperature(city string) (float64, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	resp, err := httpClient.Get(fmt.Sprintf("%s/current.json?key=%s&q=%s",
		weatherBase, apiKey, url.QueryEscape(city)))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weather api error: status %d", resp.StatusCode)
	}

	var data weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	return data.Current.TempC, nil
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	cep := r.URL.Path[1:]

	if !cepRegex.MatchString(cep) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("invalid zipcode"))
		return
	}

	city, err := getCityByCEP(cep)
	if err != nil {
		log.Printf("getCityByCEP error for cep %s: %v", cep, err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("can not find zipcode"))
		return
	}

	tempC, err := getTemperature(city)
	if err != nil {
		log.Printf("getTemperature error for city %s: %v", city, err)
		http.Error(w, "failed to get temperature", http.StatusInternalServerError)
		return
	}

	output := WeatherOutput{
		TempC: tempC,
		TempF: tempC*1.8 + 32,
		TempK: tempC + 273,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Write([]byte("rodando"))
			return
		}
		weatherHandler(w, r)
	})
	log.Printf("Server running on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
