package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abhijeet1999/weather/Producer/kafka"
	"github.com/abhijeet1999/weather/Producer/utils"
	"github.com/abhijeet1999/weather/Producer/weather"
	"github.com/abhijeet1999/weather/models"
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

		// Process based on days requirement
		if req.Days >= 4 {
			// For 4+ days: Send hourly data for 48 hours + daily data for remaining days
			err = processExtendedWeatherData(weatherService, producer, req)
		} else {
			// For 1-3 days: Use existing logic
			err = processStandardWeatherData(weatherService, producer, req)
		}

		if err != nil {
			log.Printf("❌ Failed to process weather for %s: %v", req.ZipCode, err)
			continue
		}

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("✅ Initial batch processing completed")
}

// processExtendedWeatherData handles 4+ days with hourly data for first 48 hours
func processExtendedWeatherData(weatherService *weather.WeatherService, producer *kafka.KafkaProducer, req models.WeatherRequest) error {
	log.Printf("🕐 Processing extended weather data for %s (%d days)", req.ZipCode, req.Days)

	// Get current weather
	currentWeather, err := weatherService.GetWeatherByZip(req.ZipCode, "US", "metric")
	if err != nil {
		return err
	}

	// Send current weather
	err = producer.SendCurrentWeather(req.ZipCode, currentWeather.Name, "US", currentWeather)
	if err != nil {
		log.Printf("❌ Failed to send current weather for %s: %v", req.ZipCode, err)
	}

	// Get 5-day forecast for hourly data
	forecast, err := weatherService.GetForecastByZip(req.ZipCode, "US", "metric")
	if err != nil {
		return err
	}

	// Send hourly data for first 48 hours (every 3 hours = 16 data points)
	hourlyCount := 0
	for _, item := range forecast.List {
		if hourlyCount >= 16 { // 48 hours / 3 hours per forecast = 16 items
			break
		}

		// Send individual forecast item as hourly data
		err = producer.SendHourlyWeather(req.ZipCode, forecast.City.Name, "US", item)
		if err != nil {
			log.Printf("❌ Failed to send hourly weather for %s: %v", req.ZipCode, err)
		}

		hourlyCount++
		log.Printf("📊 Sent hourly data %d/16 for %s: %.1f°C", hourlyCount, req.ZipCode, item.Main.Temp)

		// Small delay between hourly sends
		time.Sleep(50 * time.Millisecond)
	}

	// Send daily summaries for remaining days (3rd and 4th day)
	if req.Days >= 3 {
		err = producer.SendDailyWeather(req.ZipCode, forecast.City.Name, "US", forecast, 3)
		if err != nil {
			log.Printf("❌ Failed to send daily weather (day 3) for %s: %v", req.ZipCode, err)
		}
	}

	if req.Days >= 4 {
		err = producer.SendDailyWeather(req.ZipCode, forecast.City.Name, "US", forecast, 4)
		if err != nil {
			log.Printf("❌ Failed to send daily weather (day 4) for %s: %v", req.ZipCode, err)
		}
	}

	log.Printf("✅ Extended weather data completed for %s: %d hourly + %d daily",
		req.ZipCode, hourlyCount, req.Days-2)

	return nil
}

// processStandardWeatherData handles 1-3 days with existing logic
func processStandardWeatherData(weatherService *weather.WeatherService, producer *kafka.KafkaProducer, req models.WeatherRequest) error {
	// Fetch current weather
	weather, err := weatherService.GetWeatherByZip(req.ZipCode, "US", "metric")
	if err != nil {
		return err
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
			return err
		}

		// Send forecast to Kafka
		err = producer.SendForecastWeather(req.ZipCode, forecast.City.Name, "US", forecast)
		if err != nil {
			log.Printf("❌ Failed to send forecast to Kafka for %s: %v", req.ZipCode, err)
		}
	}

	return nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
