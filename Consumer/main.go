package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/abhijeet1999/weather/Consumer/api"
	"github.com/abhijeet1999/weather/Consumer/kafka"
)

func main() {
	// Configuration from environment variables
	kafkaServers := getEnvOrDefault("KAFKA_SERVERS", "kafka:9092")
	kafkaTopic := getEnvOrDefault("KAFKA_TOPIC", "weather_data")
	consumerGroupID := getEnvOrDefault("CONSUMER_GROUP_ID", "weather-consumer-group")
	metricsPort := getEnvOrDefault("METRICS_PORT", "8080")
	apiPort := getEnvOrDefault("API_PORT", "8081")

	log.Println("🚀 Starting Weather Consumer...")
	log.Printf("📥 Kafka Servers: %s", kafkaServers)
	log.Printf("📥 Kafka Topic: %s", kafkaTopic)
	log.Printf("📥 Consumer Group: %s", consumerGroupID)
	log.Printf("📊 Metrics Port: %s", metricsPort)
	log.Printf("🌐 API Port: %s", apiPort)

	// Initialize Kafka consumer
	consumer, err := kafka.NewKafkaConsumer(kafkaServers, kafkaTopic, consumerGroupID)
	if err != nil {
		log.Fatalf("❌ Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Start Prometheus metrics server
	metrics := consumer.GetMetrics()
	metrics.StartMetricsServer(metricsPort)

	// Initialize HTTP API
	weatherAPI := api.NewWeatherAPI(consumer)

	// Start Kafka consumer in background
	go func() {
		log.Println("🔄 Starting Kafka consumer...")
		consumer.StartConsuming()
	}()

	// Start HTTP API server
	go func() {
		weatherAPI.StartServer(apiPort)
	}()

	log.Println("✅ Weather Consumer started successfully!")
	log.Printf("📊 Prometheus metrics: http://localhost:%s/metrics", metricsPort)
	log.Printf("🌐 HTTP API: http://localhost:%s", apiPort)
	log.Printf("🔍 Health check: http://localhost:%s/health", apiPort)
	log.Println("⏹️  Press Ctrl+C to stop...")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🛑 Shutting down Weather Consumer...")
	log.Println("✅ Shutdown complete")
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
