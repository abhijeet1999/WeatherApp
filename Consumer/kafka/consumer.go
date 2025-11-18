package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/abhijeet1999/weather/Consumer/alerts"
	"github.com/abhijeet1999/weather/Consumer/prometheus"
	"github.com/abhijeet1999/weather/models"
	"github.com/segmentio/kafka-go"
)

// KafkaConsumer handles consuming weather data from Kafka
type KafkaConsumer struct {
	reader         *kafka.Reader
	topic          string
	groupID        string
	metrics        *prometheus.WeatherMetrics
	alertEvaluator *alerts.AlertEvaluator
}

// NewKafkaConsumer creates a new Kafka consumer instance
func NewKafkaConsumer(bootstrapServers, topic, groupID string, alertEvaluator *alerts.AlertEvaluator) (*KafkaConsumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{bootstrapServers},
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.FirstOffset,
		Dialer: &kafka.Dialer{
			Timeout: 10 * time.Second,
		},
	})

	metrics := prometheus.NewWeatherMetrics()

	log.Printf("📥 Kafka consumer connected to %s, topic: %s, group: %s", bootstrapServers, topic, groupID)

	return &KafkaConsumer{
		reader:         reader,
		topic:          topic,
		groupID:        groupID,
		metrics:        metrics,
		alertEvaluator: alertEvaluator,
	}, nil
}

// WeatherMessage represents the message structure received from Kafka
type WeatherMessage struct {
	Timestamp   time.Time                           `json:"timestamp"`
	ZipCode     string                              `json:"zip_code"`
	City        string                              `json:"city"`
	Country     string                              `json:"country"`
	Current     *models.OpenWeatherResponse         `json:"current,omitempty"`
	Forecast    *models.OpenWeatherForecastResponse `json:"forecast,omitempty"`
	Hourly      *models.ForecastItem                `json:"hourly,omitempty"`
	Daily       *DailyWeatherData                   `json:"daily,omitempty"`
	MessageType string                              `json:"message_type"` // "current", "forecast", "hourly", "daily"
}

// DailyWeatherData represents daily weather summary
type DailyWeatherData struct {
	Day         int     `json:"day"`
	Date        string  `json:"date"`
	TempMin     float32 `json:"temp_min"`
	TempMax     float32 `json:"temp_max"`
	TempAvg     float32 `json:"temp_avg"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float32 `json:"wind_speed"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
}

// StartConsuming starts consuming messages from Kafka
func (kc *KafkaConsumer) StartConsuming() {
	log.Println("🔄 Starting Kafka consumer...")

	ctx := context.Background()
	for {
		msg, err := kc.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ Error reading message: %v", err)
			continue
		}

		// Process the message
		err = kc.processMessage(msg)
		if err != nil {
			log.Printf("❌ Error processing message: %v", err)
		}
	}
}

// processMessage processes a single Kafka message
func (kc *KafkaConsumer) processMessage(msg kafka.Message) error {
	var weatherMsg WeatherMessage

	// Deserialize JSON message
	err := json.Unmarshal(msg.Value, &weatherMsg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	log.Printf("📥 Received weather data: %s (%s) - %s",
		weatherMsg.City, weatherMsg.ZipCode, weatherMsg.MessageType)

	// Process based on message type
	switch weatherMsg.MessageType {
	case "current":
		return kc.processCurrentWeather(weatherMsg)
	case "forecast":
		return kc.processForecastWeather(weatherMsg)
	case "hourly":
		return kc.processHourlyWeather(weatherMsg)
	case "daily":
		return kc.processDailyWeather(weatherMsg)
	default:
		return fmt.Errorf("unknown message type: %s", weatherMsg.MessageType)
	}
}

// processCurrentWeather processes current weather data
func (kc *KafkaConsumer) processCurrentWeather(msg WeatherMessage) error {
	if msg.Current == nil {
		return fmt.Errorf("current weather data is nil")
	}

	// Update Prometheus metrics
	kc.metrics.UpdateCurrentWeatherMetrics(
		msg.City,
		msg.Current.Main.Temp,
		float32(msg.Current.Main.Humidity),
		msg.Current.Wind.Speed,
		float32(msg.Current.Main.Pressure),
	)

	// Evaluate alerts
	if kc.alertEvaluator != nil {
		alerts := kc.alertEvaluator.EvaluateCurrentWeather(*msg.Current, msg.ZipCode)
		kc.processAlerts(alerts)
	}

	log.Printf("📊 Updated metrics for %s: Temp=%.1f°C, Humidity=%d%%, Wind=%.1fm/s",
		msg.City, msg.Current.Main.Temp, msg.Current.Main.Humidity, msg.Current.Wind.Speed)

	return nil
}

// processForecastWeather processes forecast weather data
func (kc *KafkaConsumer) processForecastWeather(msg WeatherMessage) error {
	if msg.Forecast == nil {
		return fmt.Errorf("forecast weather data is nil")
	}

	// Process each forecast item
	for _, item := range msg.Forecast.List {
		// Update Prometheus metrics for forecast data
		kc.metrics.UpdateForecastWeatherMetrics(
			msg.City,
			item.Main.Temp,
			float32(item.Main.Humidity),
			item.Wind.Speed,
			float32(item.Main.Pressure),
			item.Dt,
		)
	}

	log.Printf("📊 Updated forecast metrics for %s: %d forecast items",
		msg.City, len(msg.Forecast.List))

	return nil
}

// processHourlyWeather processes individual hourly weather data
func (kc *KafkaConsumer) processHourlyWeather(msg WeatherMessage) error {
	if msg.Hourly == nil {
		return fmt.Errorf("hourly weather data is nil")
	}

	// Update Prometheus metrics for hourly data
	kc.metrics.UpdateForecastWeatherMetrics(
		msg.City,
		msg.Hourly.Main.Temp,
		float32(msg.Hourly.Main.Humidity),
		msg.Hourly.Wind.Speed,
		float32(msg.Hourly.Main.Pressure),
		msg.Hourly.Dt,
	)

	// Evaluate alerts for hourly data
	if kc.alertEvaluator != nil {
		alerts := kc.alertEvaluator.EvaluateHourlyWeather(*msg.Hourly, msg.ZipCode)
		kc.processAlerts(alerts)
	}

	log.Printf("📊 Updated hourly metrics for %s: Temp=%.1f°C, Humidity=%d%%, Wind=%.1fm/s",
		msg.City, msg.Hourly.Main.Temp, msg.Hourly.Main.Humidity, msg.Hourly.Wind.Speed)

	return nil
}

// processDailyWeather processes daily weather summary data
func (kc *KafkaConsumer) processDailyWeather(msg WeatherMessage) error {
	if msg.Daily == nil {
		return fmt.Errorf("daily weather data is nil")
	}

	// Update Prometheus metrics for daily data
	// Use average temperature for daily metrics
	kc.metrics.UpdateCurrentWeatherMetrics(
		msg.City,
		msg.Daily.TempAvg,
		float32(msg.Daily.Humidity),
		msg.Daily.WindSpeed,
		1013.25, // Default pressure for daily data
	)

	log.Printf("📊 Updated daily metrics for %s (Day %d): TempAvg=%.1f°C, TempMin=%.1f°C, TempMax=%.1f°C, Humidity=%d%%, Wind=%.1fm/s",
		msg.City, msg.Daily.Day, msg.Daily.TempAvg, msg.Daily.TempMin, msg.Daily.TempMax, msg.Daily.Humidity, msg.Daily.WindSpeed)

	return nil
}

// Close closes the Kafka consumer
func (kc *KafkaConsumer) Close() {
	kc.reader.Close()
}

// GetMetrics returns the Prometheus metrics instance
func (kc *KafkaConsumer) GetMetrics() *prometheus.WeatherMetrics {
	return kc.metrics
}

// processAlerts processes triggered alerts
func (kc *KafkaConsumer) processAlerts(weatherAlerts []alerts.WeatherAlert) {
	for _, alert := range weatherAlerts {
		log.Printf("🚨 ALERT [%s] %s: %s", alert.Severity, alert.Type, alert.Description)

		// Update Prometheus metrics with alert information
		kc.metrics.UpdateAlertMetrics(
			alert.City,
			alert.Type,
			alert.Severity,
			alert.Value,
			alert.Threshold,
		)
	}
}
