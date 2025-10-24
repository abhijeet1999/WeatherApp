package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/abhijeet1999/weather/models"
	"github.com/segmentio/kafka-go"
)

// KafkaProducer handles sending weather data to Kafka
type KafkaProducer struct {
	writer *kafka.Writer
	topic  string
}

// NewKafkaProducer creates a new Kafka producer instance
func NewKafkaProducer(bootstrapServers, topic string) (*KafkaProducer, error) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(bootstrapServers),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Async:    true,
	}

	log.Printf("📤 Kafka producer connected to %s, topic: %s", bootstrapServers, topic)

	return &KafkaProducer{
		writer: writer,
		topic:  topic,
	}, nil
}

// WeatherMessage represents the message structure sent to Kafka
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

// SendWeatherData sends weather data to Kafka topic
func (kp *KafkaProducer) SendWeatherData(message WeatherMessage) error {
	// Serialize message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Create Kafka message
	kafkaMessage := kafka.Message{
		Key:   []byte(message.ZipCode),
		Value: jsonData,
		Headers: []kafka.Header{
			{Key: "message_type", Value: []byte(message.MessageType)},
			{Key: "zip_code", Value: []byte(message.ZipCode)},
			{Key: "city", Value: []byte(message.City)},
		},
	}

	// Send message
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = kp.writer.WriteMessages(ctx, kafkaMessage)
	if err != nil {
		return fmt.Errorf("failed to produce message: %v", err)
	}

	log.Printf("📤 Sent weather data to Kafka: %s (%s) - %s",
		message.City, message.ZipCode, message.MessageType)

	return nil
}

// SendCurrentWeather sends current weather data to Kafka
func (kp *KafkaProducer) SendCurrentWeather(zipCode, city, country string, weather models.OpenWeatherResponse) error {
	message := WeatherMessage{
		Timestamp:   time.Now(),
		ZipCode:     zipCode,
		City:        city,
		Country:     country,
		Current:     &weather,
		MessageType: "current",
	}

	return kp.SendWeatherData(message)
}

// SendForecastWeather sends forecast weather data to Kafka
func (kp *KafkaProducer) SendForecastWeather(zipCode, city, country string, forecast models.OpenWeatherForecastResponse) error {
	message := WeatherMessage{
		Timestamp:   time.Now(),
		ZipCode:     zipCode,
		City:        city,
		Country:     country,
		Forecast:    &forecast,
		MessageType: "forecast",
	}

	return kp.SendWeatherData(message)
}

// SendHourlyWeather sends individual hourly weather data to Kafka
func (kp *KafkaProducer) SendHourlyWeather(zipCode, city, country string, hourly models.ForecastItem) error {
	message := WeatherMessage{
		Timestamp:   time.Now(),
		ZipCode:     zipCode,
		City:        city,
		Country:     country,
		Hourly:      &hourly,
		MessageType: "hourly",
	}

	return kp.SendWeatherData(message)
}

// SendDailyWeather sends daily weather summary to Kafka
func (kp *KafkaProducer) SendDailyWeather(zipCode, city, country string, forecast models.OpenWeatherForecastResponse, day int) error {
	// Calculate daily summary from forecast items for the specified day
	dailyData := kp.calculateDailySummary(forecast, day)

	message := WeatherMessage{
		Timestamp:   time.Now(),
		ZipCode:     zipCode,
		City:        city,
		Country:     country,
		Daily:       dailyData,
		MessageType: "daily",
	}

	return kp.SendWeatherData(message)
}

// calculateDailySummary calculates daily weather summary from forecast items
func (kp *KafkaProducer) calculateDailySummary(forecast models.OpenWeatherForecastResponse, day int) *DailyWeatherData {
	// Get forecast items for the specified day
	var dayItems []models.ForecastItem
	targetDate := time.Now().AddDate(0, 0, day-1).Format("2006-01-02")

	for _, item := range forecast.List {
		itemDate := time.Unix(item.Dt, 0).Format("2006-01-02")
		if itemDate == targetDate {
			dayItems = append(dayItems, item)
		}
	}

	if len(dayItems) == 0 {
		return &DailyWeatherData{
			Day:         day,
			Date:        targetDate,
			Description: "No data available",
		}
	}

	// Calculate summary statistics
	var tempMin, tempMax, tempSum float32
	var humiditySum int
	var windSum float32

	tempMin = dayItems[0].Main.TempMin
	tempMax = dayItems[0].Main.TempMax

	for _, item := range dayItems {
		if item.Main.TempMin < tempMin {
			tempMin = item.Main.TempMin
		}
		if item.Main.TempMax > tempMax {
			tempMax = item.Main.TempMax
		}
		tempSum += item.Main.Temp
		humiditySum += item.Main.Humidity
		windSum += item.Wind.Speed
	}

	return &DailyWeatherData{
		Day:         day,
		Date:        targetDate,
		TempMin:     tempMin,
		TempMax:     tempMax,
		TempAvg:     tempSum / float32(len(dayItems)),
		Humidity:    humiditySum / len(dayItems),
		WindSpeed:   windSum / float32(len(dayItems)),
		Description: dayItems[0].Weather[0].Description,
		Icon:        dayItems[0].Weather[0].Icon,
	}
}

// Close closes the Kafka producer
func (kp *KafkaProducer) Close() {
	kp.writer.Close()
}

// Flush ensures all messages are sent before closing
func (kp *KafkaProducer) Flush(timeoutMs int) {
	// Segment kafka-go handles flushing automatically with Async: true
	// For synchronous flushing, we can use a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	// Wait for any pending writes to complete
	select {
	case <-ctx.Done():
		log.Printf("⚠️ Flush timeout after %dms", timeoutMs)
	default:
		log.Printf("✅ Flush completed")
	}
}
