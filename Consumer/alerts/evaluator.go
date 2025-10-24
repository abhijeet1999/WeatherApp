package alerts

import (
	"fmt"
	"log"
	"time"

	"github.com/abhijeet1999/weather/models"
)

// AlertEvaluator handles evaluating weather conditions and triggering alerts
type AlertEvaluator struct {
	alertRules map[string]AlertRule
}

// AlertRule defines alert conditions for a specific location
type AlertRule struct {
	ZipCode       string  `json:"zip_code"`
	City          string  `json:"city"`
	AlertTemp     float32 `json:"alert_temp"`
	HighTempAlert float32 `json:"high_temp_alert"`
	LowTempAlert  float32 `json:"low_temp_alert"`
	WindAlert     float32 `json:"wind_alert"`
	HumidityAlert int     `json:"humidity_alert"`
	PressureAlert int     `json:"pressure_alert"`
}

// WeatherAlert represents a triggered alert
type WeatherAlert struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	City        string    `json:"city"`
	ZipCode     string    `json:"zip_code"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

// NewAlertEvaluator creates a new alert evaluator
func NewAlertEvaluator() *AlertEvaluator {
	return &AlertEvaluator{
		alertRules: make(map[string]AlertRule),
	}
}

// AddAlertRule adds an alert rule for a specific location
func (ae *AlertEvaluator) AddAlertRule(zipCode, city string, alertTemp, alertWind float32, alertHumidity int) {
	rule := AlertRule{
		ZipCode:       zipCode,
		City:          city,
		AlertTemp:     alertTemp,
		HighTempAlert: alertTemp + 10, // 10°C above custom threshold
		LowTempAlert:  alertTemp - 5,  // 5°C below custom threshold
		WindAlert:     alertWind,      // User-specified wind threshold
		HumidityAlert: alertHumidity,  // User-specified humidity threshold
		PressureAlert: 1000,           // 1000 hPa pressure (fixed)
	}

	ae.alertRules[zipCode] = rule
	log.Printf("📋 Added alert rule for %s (%s): Temp=%.1f°C, Wind=%.1fm/s, Humidity=%d%%",
		city, zipCode, alertTemp, alertWind, alertHumidity)
}

// EvaluateCurrentWeather evaluates current weather conditions and returns alerts
func (ae *AlertEvaluator) EvaluateCurrentWeather(weather models.OpenWeatherResponse, zipCode string) []WeatherAlert {
	var alerts []WeatherAlert

	rule, exists := ae.alertRules[zipCode]
	if !exists {
		log.Printf("⚠️ No alert rule found for zip code: %s", zipCode)
		return alerts
	}

	// Temperature alerts
	alerts = append(alerts, ae.evaluateTemperatureAlerts(weather, rule)...)

	// Wind alerts
	alerts = append(alerts, ae.evaluateWindAlerts(weather, rule)...)

	// Humidity alerts
	alerts = append(alerts, ae.evaluateHumidityAlerts(weather, rule)...)

	// Pressure alerts
	alerts = append(alerts, ae.evaluatePressureAlerts(weather, rule)...)

	// Weather condition alerts
	alerts = append(alerts, ae.evaluateWeatherConditionAlerts(weather, rule)...)

	return alerts
}

// EvaluateHourlyWeather evaluates hourly weather conditions and returns alerts
func (ae *AlertEvaluator) EvaluateHourlyWeather(hourly models.ForecastItem, zipCode string) []WeatherAlert {
	var alerts []WeatherAlert

	_, exists := ae.alertRules[zipCode]
	if !exists {
		return alerts
	}

	// Convert ForecastItem to OpenWeatherResponse for evaluation
	weather := models.OpenWeatherResponse{
		Main: struct {
			Temp      float32 `json:"temp"`
			FeelsLike float32 `json:"feels_like"`
			TempMin   float32 `json:"temp_min"`
			TempMax   float32 `json:"temp_max"`
			Pressure  int     `json:"pressure"`
			Humidity  int     `json:"humidity"`
		}{
			Temp:     hourly.Main.Temp,
			Pressure: hourly.Main.Pressure,
			Humidity: hourly.Main.Humidity,
		},
		Wind: struct {
			Speed float32 `json:"speed"`
			Deg   int     `json:"deg"`
		}{
			Speed: hourly.Wind.Speed,
		},
		Weather: hourly.Weather,
	}

	return ae.EvaluateCurrentWeather(weather, zipCode)
}

// evaluateTemperatureAlerts checks for temperature-related alerts
func (ae *AlertEvaluator) evaluateTemperatureAlerts(weather models.OpenWeatherResponse, rule AlertRule) []WeatherAlert {
	var alerts []WeatherAlert

	temp := weather.Main.Temp

	// Custom temperature alert (from input.txt)
	if temp >= rule.AlertTemp {
		alerts = append(alerts, WeatherAlert{
			Type:        "temperature",
			Severity:    "warning",
			Message:     fmt.Sprintf("Temperature reached custom threshold"),
			City:        rule.City,
			ZipCode:     rule.ZipCode,
			Value:       float64(temp),
			Threshold:   float64(rule.AlertTemp),
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Temperature in %s is %.1f°C, which has reached the custom threshold of %.1f°C", rule.City, temp, rule.AlertTemp),
		})
	}

	// High temperature alert
	if temp >= rule.HighTempAlert {
		alerts = append(alerts, WeatherAlert{
			Type:        "high_temperature",
			Severity:    "critical",
			Message:     fmt.Sprintf("High temperature detected"),
			City:        rule.City,
			ZipCode:     rule.ZipCode,
			Value:       float64(temp),
			Threshold:   float64(rule.HighTempAlert),
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Temperature in %s is %.1f°C, which is above the high temperature threshold of %.1f°C", rule.City, temp, rule.HighTempAlert),
		})
	}

	// Low temperature alert
	if temp <= rule.LowTempAlert {
		alerts = append(alerts, WeatherAlert{
			Type:        "low_temperature",
			Severity:    "warning",
			Message:     fmt.Sprintf("Low temperature detected"),
			City:        rule.City,
			ZipCode:     rule.ZipCode,
			Value:       float64(temp),
			Threshold:   float64(rule.LowTempAlert),
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Temperature in %s is %.1f°C, which is below the low temperature threshold of %.1f°C", rule.City, temp, rule.LowTempAlert),
		})
	}

	return alerts
}

// evaluateWindAlerts checks for wind-related alerts
func (ae *AlertEvaluator) evaluateWindAlerts(weather models.OpenWeatherResponse, rule AlertRule) []WeatherAlert {
	var alerts []WeatherAlert

	windSpeed := weather.Wind.Speed

	if windSpeed >= rule.WindAlert {
		alerts = append(alerts, WeatherAlert{
			Type:        "high_wind",
			Severity:    "warning",
			Message:     fmt.Sprintf("High wind speed detected"),
			City:        rule.City,
			ZipCode:     rule.ZipCode,
			Value:       float64(windSpeed),
			Threshold:   float64(rule.WindAlert),
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Wind speed in %s is %.1f m/s, which is above the threshold of %.1f m/s", rule.City, windSpeed, rule.WindAlert),
		})
	}

	return alerts
}

// evaluateHumidityAlerts checks for humidity-related alerts
func (ae *AlertEvaluator) evaluateHumidityAlerts(weather models.OpenWeatherResponse, rule AlertRule) []WeatherAlert {
	var alerts []WeatherAlert

	humidity := weather.Main.Humidity

	if humidity >= rule.HumidityAlert {
		alerts = append(alerts, WeatherAlert{
			Type:        "high_humidity",
			Severity:    "warning",
			Message:     fmt.Sprintf("High humidity detected"),
			City:        rule.City,
			ZipCode:     rule.ZipCode,
			Value:       float64(humidity),
			Threshold:   float64(rule.HumidityAlert),
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Humidity in %s is %d%%, which is above the threshold of %d%%", rule.City, humidity, rule.HumidityAlert),
		})
	}

	return alerts
}

// evaluatePressureAlerts checks for pressure-related alerts
func (ae *AlertEvaluator) evaluatePressureAlerts(weather models.OpenWeatherResponse, rule AlertRule) []WeatherAlert {
	var alerts []WeatherAlert

	pressure := weather.Main.Pressure

	if pressure <= rule.PressureAlert {
		alerts = append(alerts, WeatherAlert{
			Type:        "low_pressure",
			Severity:    "warning",
			Message:     fmt.Sprintf("Low atmospheric pressure detected"),
			City:        rule.City,
			ZipCode:     rule.ZipCode,
			Value:       float64(pressure),
			Threshold:   float64(rule.PressureAlert),
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Atmospheric pressure in %s is %d hPa, which is below the threshold of %d hPa", rule.City, pressure, rule.PressureAlert),
		})
	}

	return alerts
}

// evaluateWeatherConditionAlerts checks for weather condition alerts
func (ae *AlertEvaluator) evaluateWeatherConditionAlerts(weather models.OpenWeatherResponse, rule AlertRule) []WeatherAlert {
	var alerts []WeatherAlert

	if len(weather.Weather) > 0 {
		condition := weather.Weather[0].Main
		description := weather.Weather[0].Description

		// Severe weather conditions
		severeConditions := map[string]string{
			"Thunderstorm": "critical",
			"Snow":         "warning",
			"Rain":         "warning",
			"Drizzle":      "info",
		}

		if severity, exists := severeConditions[condition]; exists {
			alerts = append(alerts, WeatherAlert{
				Type:        "weather_condition",
				Severity:    severity,
				Message:     fmt.Sprintf("Severe weather condition: %s", condition),
				City:        rule.City,
				ZipCode:     rule.ZipCode,
				Value:       0,
				Threshold:   0,
				Timestamp:   time.Now(),
				Description: fmt.Sprintf("Weather condition in %s: %s (%s)", rule.City, condition, description),
			})
		}
	}

	return alerts
}

// GetAlertRules returns all configured alert rules
func (ae *AlertEvaluator) GetAlertRules() map[string]AlertRule {
	return ae.alertRules
}

// GetAlertRule returns alert rule for a specific zip code
func (ae *AlertEvaluator) GetAlertRule(zipCode string) (AlertRule, bool) {
	rule, exists := ae.alertRules[zipCode]
	return rule, exists
}
