package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/abhijeet1999/weather/models"
	"github.com/abhijeet1999/weather/Producer/utils"
)

// WeatherService handles weather-related API calls
type WeatherService struct {
	httpClient http.Client
}

// NewWeatherService creates a new WeatherService instance
func NewWeatherService() *WeatherService {
	return &WeatherService{
		httpClient: http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// GetWeather fetches weather data by latitude and longitude
func (ws *WeatherService) GetWeather(lat, lon float64, units string) (models.OpenWeatherResponse, error) {
	var weather models.OpenWeatherResponse

	u := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=%s",
		lat, lon, utils.GetOpenWeatherMapApiKey(), units,
	)

	r, err := ws.httpClient.Get(u)
	if err != nil {
		return weather, err
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return weather, fmt.Errorf("OpenWeatherRequest Failed: %s", r.Status)
	}

	err = json.NewDecoder(r.Body).Decode(&weather)
	return weather, err
}

// GetLatLon converts ZIP code to latitude and longitude coordinates
func (ws *WeatherService) GetLatLon(zip, country string) (float64, float64, error) {
	var geo models.GeoResponse

	url := fmt.Sprintf(
		"http://api.openweathermap.org/geo/1.0/zip?zip=%s,%s&appid=%s",
		zip, country, utils.GetOpenWeatherMapApiKey(),
	)

	resp, err := ws.httpClient.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, 0, fmt.Errorf("failed to get location: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&geo)
	if err != nil {
		return 0, 0, err
	}

	return geo.Lat, geo.Lon, nil
}

// GetWeatherByZip fetches weather data by ZIP code and country
func (ws *WeatherService) GetWeatherByZip(zip, country, units string) (models.OpenWeatherResponse, error) {
	lat, lon, err := ws.GetLatLon(zip, country)
	if err != nil {
		return models.OpenWeatherResponse{}, err
	}

	return ws.GetWeather(lat, lon, units)
}

// GetForecast fetches 5-day weather forecast by latitude and longitude
func (ws *WeatherService) GetForecast(lat, lon float64, units string) (models.OpenWeatherForecastResponse, error) {
	var forecast models.OpenWeatherForecastResponse

	u := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&appid=%s&units=%s",
		lat, lon, utils.GetOpenWeatherMapApiKey(), units,
	)

	r, err := ws.httpClient.Get(u)
	if err != nil {
		return forecast, err
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return forecast, fmt.Errorf("OpenWeatherForecastRequest Failed: %s", r.Status)
	}

	err = json.NewDecoder(r.Body).Decode(&forecast)
	return forecast, err
}

// GetForecastByZip fetches 5-day weather forecast by ZIP code and country
func (ws *WeatherService) GetForecastByZip(zip, country, units string) (models.OpenWeatherForecastResponse, error) {
	lat, lon, err := ws.GetLatLon(zip, country)
	if err != nil {
		return models.OpenWeatherForecastResponse{}, err
	}

	return ws.GetForecast(lat, lon, units)
}
