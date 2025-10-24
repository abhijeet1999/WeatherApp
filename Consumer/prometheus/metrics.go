package prometheus

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// WeatherMetrics holds all Prometheus metrics for weather data
type WeatherMetrics struct {
	// Current weather metrics
	temperatureCelsius *prometheus.GaugeVec
	humidityPercent    *prometheus.GaugeVec
	windSpeedMps       *prometheus.GaugeVec
	pressureHpa        *prometheus.GaugeVec

	// Forecast weather metrics
	forecastTemperature *prometheus.GaugeVec
	forecastHumidity    *prometheus.GaugeVec
	forecastWindSpeed   *prometheus.GaugeVec
	forecastPressure    *prometheus.GaugeVec

	// Counter metrics
	weatherRequestsTotal *prometheus.CounterVec
	weatherErrorsTotal   *prometheus.CounterVec

	// Histogram metrics
	weatherProcessingTime *prometheus.HistogramVec
}

// NewWeatherMetrics creates a new WeatherMetrics instance
func NewWeatherMetrics() *WeatherMetrics {
	metrics := &WeatherMetrics{
		// Current weather gauges
		temperatureCelsius: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "weather_temperature_celsius",
				Help: "Current temperature in Celsius",
			},
			[]string{"city", "zip_code"},
		),

		humidityPercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "weather_humidity_percent",
				Help: "Current humidity percentage",
			},
			[]string{"city", "zip_code"},
		),

		windSpeedMps: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "weather_wind_speed_mps",
				Help: "Current wind speed in meters per second",
			},
			[]string{"city", "zip_code"},
		),

		pressureHpa: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "weather_pressure_hpa",
				Help: "Current atmospheric pressure in hPa",
			},
			[]string{"city", "zip_code"},
		),

		// Forecast weather gauges
		forecastTemperature: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "weather_forecast_temperature_celsius",
				Help: "Forecast temperature in Celsius",
			},
			[]string{"city", "zip_code", "forecast_time"},
		),

		forecastHumidity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "weather_forecast_humidity_percent",
				Help: "Forecast humidity percentage",
			},
			[]string{"city", "zip_code", "forecast_time"},
		),

		forecastWindSpeed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "weather_forecast_wind_speed_mps",
				Help: "Forecast wind speed in meters per second",
			},
			[]string{"city", "zip_code", "forecast_time"},
		),

		forecastPressure: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "weather_forecast_pressure_hpa",
				Help: "Forecast atmospheric pressure in hPa",
			},
			[]string{"city", "zip_code", "forecast_time"},
		),

		// Counter metrics
		weatherRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_requests_total",
				Help: "Total number of weather requests processed",
			},
			[]string{"city", "zip_code", "status"},
		),

		weatherErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_errors_total",
				Help: "Total number of weather processing errors",
			},
			[]string{"city", "zip_code", "error_type"},
		),

		// Histogram metrics
		weatherProcessingTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "weather_processing_duration_seconds",
				Help:    "Time spent processing weather requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"city", "zip_code"},
		),
	}

	// Register all metrics
	prometheus.MustRegister(
		metrics.temperatureCelsius,
		metrics.humidityPercent,
		metrics.windSpeedMps,
		metrics.pressureHpa,
		metrics.forecastTemperature,
		metrics.forecastHumidity,
		metrics.forecastWindSpeed,
		metrics.forecastPressure,
		metrics.weatherRequestsTotal,
		metrics.weatherErrorsTotal,
		metrics.weatherProcessingTime,
	)

	return metrics
}

// UpdateCurrentWeatherMetrics updates current weather metrics
func (wm *WeatherMetrics) UpdateCurrentWeatherMetrics(city string, temp, humidity, windSpeed, pressure float32) {
	zipCode := "unknown" // We could extract this from the message if needed

	wm.temperatureCelsius.WithLabelValues(city, zipCode).Set(float64(temp))
	wm.humidityPercent.WithLabelValues(city, zipCode).Set(float64(humidity))
	wm.windSpeedMps.WithLabelValues(city, zipCode).Set(float64(windSpeed))
	wm.pressureHpa.WithLabelValues(city, zipCode).Set(float64(pressure))
}

// UpdateForecastWeatherMetrics updates forecast weather metrics
func (wm *WeatherMetrics) UpdateForecastWeatherMetrics(city string, temp, humidity, windSpeed, pressure float32, timestamp int64) {
	zipCode := "unknown"
	forecastTime := time.Unix(timestamp, 0).Format("2006-01-02T15:04:05")

	wm.forecastTemperature.WithLabelValues(city, zipCode, forecastTime).Set(float64(temp))
	wm.forecastHumidity.WithLabelValues(city, zipCode, forecastTime).Set(float64(humidity))
	wm.forecastWindSpeed.WithLabelValues(city, zipCode, forecastTime).Set(float64(windSpeed))
	wm.forecastPressure.WithLabelValues(city, zipCode, forecastTime).Set(float64(pressure))
}

// IncrementWeatherRequests increments the weather requests counter
func (wm *WeatherMetrics) IncrementWeatherRequests(city, zipCode, status string) {
	wm.weatherRequestsTotal.WithLabelValues(city, zipCode, status).Inc()
}

// IncrementWeatherErrors increments the weather errors counter
func (wm *WeatherMetrics) IncrementWeatherErrors(city, zipCode, errorType string) {
	wm.weatherErrorsTotal.WithLabelValues(city, zipCode, errorType).Inc()
}

// RecordProcessingTime records the processing time for a weather request
func (wm *WeatherMetrics) RecordProcessingTime(city, zipCode string, duration time.Duration) {
	wm.weatherProcessingTime.WithLabelValues(city, zipCode).Observe(duration.Seconds())
}

// SetTestTemperature sets a test temperature for alerting purposes
func (wm *WeatherMetrics) SetTestTemperature(city string, temperature float64) {
	zipCode := "test"
	wm.temperatureCelsius.WithLabelValues(city, zipCode).Set(temperature)
	log.Printf("🧪 Set test temperature for %s: %.1f°C", city, temperature)
}

// StartMetricsServer starts the Prometheus metrics HTTP server
func (wm *WeatherMetrics) StartMetricsServer(port string) {
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Printf("📊 Starting Prometheus metrics server on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("❌ Error starting metrics server: %v", err)
		}
	}()
}
