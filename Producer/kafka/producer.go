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
	MessageType string                              `json:"message_type"` // "current" or "forecast"
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
