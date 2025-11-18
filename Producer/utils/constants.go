package utils

import "os"

// GetOpenWeatherMapApiKey retrieves the API key from environment variable
func GetOpenWeatherMapApiKey() string {
	return os.Getenv("WEATHER_API_KEY")
}

const (
	WeatherPeriodCurrent = "current"
	WeatherPeriodMinutly = "minutely"
	WeatherPeriodDaily   = "daily"
	WeatherPeriodHourly  = "hourly"
	UnitImperial         = "imperial"
	UnitMetric           = "metric"
)
