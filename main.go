package main

import (
	"encoding/json" // Dekodowanie odpowiedzi JSON z API pogodowego
	"fmt"
	"html/template" // Silnik szablonów HTML
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const port = "8089"

type GeoResponse struct {
	Results []GeoResult `json:"results"`
}

type GeoResult struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type WeatherResponse struct {
	CurrentWeather CurrentWeather `json:"current_weather"`
}

type CurrentWeather struct {
	Temperature float64 `json:"temperature"`
	WindSpeed   float64 `json:"windspeed"`
	Weathercode int     `json:"weathercode"`
}

type TemplateData struct {
	Cities  []string
	Weather *WeatherInfo
	Error   string
}

type WeatherInfo struct {
	City        string
	Country     string
	Temperature float64
	WindSpeed   float64
	Description string
}

var cities = []string{
	"Warszawa",
	"Kraków",
	"Wrocław",
	"Gdańsk",
	"Poznań",
	"Kyiv",
	"Lviv",
	"Berlin",
	"Paris",
	"London",
	"Rome",
	"Madrid",
	"Prague",
	"Vienna",
	"Amsterdam",
}

func weatherDescription(code int) string {
	switch {
	case code == 0:
		return "Bezchmurnie"
	case code <= 3:
		return "Częściowe zachmurzenie"
	case code <= 49:
		return "Mgła"
	case code <= 59:
		return "Mżawka"
	case code <= 69:
		return "Deszcz"
	case code <= 79:
		return "Śnieg"
	case code <= 82:
		return "Deszcz z przerwami"
	case code <= 86:
		return "Śnieg z przerwami"
	default:
		return "Burza"
	}
}

func getCoordinates(city string) (*GeoResult, error) {
	apiURL := fmt.Sprintf(
		"https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=pl&format=json",
		url.QueryEscape(city),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("błąd połączenia z API geokodowania: %w", err)
	}
	defer resp.Body.Close()

	var geoResp GeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return nil, fmt.Errorf("błąd dekodowania odpowiedzi: %w", err)
	}

	if len(geoResp.Results) == 0 {
		return nil, fmt.Errorf("nie znaleziono miasta: %s", city)
	}

	return &geoResp.Results[0], nil
}

func getWeather(lat, lon float64) (*WeatherResponse, error) {
	apiURL := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current_weather=true",
		lat, lon,
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("błąd połączenia z API pogodowym: %w", err)
	}
	defer resp.Body.Close()

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("błąd dekodowania odpowiedzi pogodowej: %w", err)
	}

	return &weatherResp, nil
}

func handleIndex(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := TemplateData{
			Cities: cities,
		}

		if r.Method == http.MethodPost {
			city := r.FormValue("city")

			geo, err := getCoordinates(city)
			if err != nil {
				data.Error = err.Error()
				tmpl.Execute(w, data)
				return
			}

			weather, err := getWeather(geo.Latitude, geo.Longitude)
			if err != nil {
				data.Error = err.Error()
				tmpl.Execute(w, data)
				return
			}

			data.Weather = &WeatherInfo{
				City:        geo.Name,
				Country:     geo.Country,
				Temperature: weather.CurrentWeather.Temperature,
				WindSpeed:   weather.CurrentWeather.WindSpeed,
				Description: weatherDescription(weather.CurrentWeather.Weathercode),
			}
		}

		tmpl.Execute(w, data)
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-health-check" {
		resp, err := http.Get("http://localhost:" + port + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	log.Printf("=== Aplikacja pogodowa ===")
	log.Printf("Data uruchomienia: %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("Autor: Nazarii Loboda")
	log.Printf("Nasłuchiwanie na porcie TCP: %s", port)

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Błąd wczytywania szablonu: %v", err)
	}

	http.HandleFunc("/", handleIndex(tmpl))

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	log.Printf("Serwer uruchomiony: http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Błąd uruchomienia serwera: %v", err)
	}
}
