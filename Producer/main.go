package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abhijeet1999/weather/Producer/kafka"
	"github.com/abhijeet1999/weather/Producer/weather"
	"github.com/abhijeet1999/weather/Producer/utils"
)

func main() {
	// Check for required environment variables
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ WEATHER_API_KEY environment variable is required")
	}

	// Configuration from environment variables
	kafkaServers := getEnvOrDefault("KAFKA_SERVERS", "kafka:9092")
	kafkaTopic := getEnvOrDefault("KAFKA_TOPIC", "weather_data")
	inputFile := getEnvOrDefault("INPUT_FILE", "input.txt")

	log.Println("🚀 Starting Weather Producer...")
	log.Printf("📤 Kafka Servers: %s", kafkaServers)
	log.Printf("📤 Kafka Topic: %s", kafkaTopic)
	log.Printf("📄 Input File: %s", inputFile)

	// Initialize weather service
	weatherService := weather.NewWeatherService()

	// Initialize Kafka producer
	producer, err := kafka.NewKafkaProducer(kafkaServers, kafkaTopic)
	if err != nil {
		log.Fatalf("❌ Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Process initial batch from input file
	go func() {
		time.Sleep(2 * time.Second) // Wait for Kafka to be ready
		processInitialBatch(weatherService, producer, inputFile)
	}()

	log.Println("✅ Weather Producer started successfully!")
	log.Println("⏹️  Press Ctrl+C to stop...")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🛑 Shutting down Weather Producer...")
	producer.Flush(1000) // Flush remaining messages
	log.Println("✅ Shutdown complete")
}

// processInitialBatch processes the initial batch from input file
func processInitialBatch(weatherService *weather.WeatherService, producer *kafka.KafkaProducer, inputFile string) {
	log.Printf("📋 Processing initial batch from %s...", inputFile)

	// Parse input file
	requests, err := utils.ParseInputFile(inputFile)
	if err != nil {
		log.Printf("❌ Error parsing input file: %v", err)
		return
	}

	log.Printf("🚀 Processing %d weather requests...", len(requests))

	// Process each request
	for i, req := range requests {
		log.Printf("📤 Processing request %d: %s (%d days)", i+1, req.ZipCode, req.Days)

		// Fetch current weather
		weather, err := weatherService.GetWeatherByZip(req.ZipCode, "US", "metric")
		if err != nil {
			log.Printf("❌ Failed to fetch weather for %s: %v", req.ZipCode, err)
			continue
		}

		// Send current weather to Kafka
		err = producer.SendCurrentWeather(req.ZipCode, weather.Name, "US", weather)
		if err != nil {
			log.Printf("❌ Failed to send current weather to Kafka for %s: %v", req.ZipCode, err)
		}

		// Fetch forecast if requested
		if req.Days > 0 {
			forecast, err := weatherService.GetForecastByZip(req.ZipCode, "US", "metric")
			if err != nil {
				log.Printf("❌ Failed to fetch forecast for %s: %v", req.ZipCode, err)
				continue
			}

			// Send forecast to Kafka
			err = producer.SendForecastWeather(req.ZipCode, forecast.City.Name, "US", forecast)
			if err != nil {
				log.Printf("❌ Failed to send forecast to Kafka for %s: %v", req.ZipCode, err)
			}
		}

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("✅ Initial batch processing completed")
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
