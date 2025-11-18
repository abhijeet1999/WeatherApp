package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/abhijeet1999/weather/Consumer/kafka"
	"github.com/gorilla/mux"
)

// WeatherAPI handles HTTP requests for weather data
type WeatherAPI struct {
	consumer *kafka.KafkaConsumer
	router   *mux.Router
}

// WeatherResponse represents the API response structure
type WeatherResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewWeatherAPI creates a new WeatherAPI instance
func NewWeatherAPI(consumer *kafka.KafkaConsumer) *WeatherAPI {
	api := &WeatherAPI{
		consumer: consumer,
		router:   mux.NewRouter(),
	}

	api.setupRoutes()
	return api
}

// setupRoutes configures all API routes
func (api *WeatherAPI) setupRoutes() {
	api.router.HandleFunc("/health", api.healthCheck).Methods("GET")
	api.router.HandleFunc("/metrics", api.getMetrics).Methods("GET")
	api.router.HandleFunc("/test/temperature", api.setTestTemperature).Methods("POST")
}

// healthCheck returns the health status of the API
func (api *WeatherAPI) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := WeatherResponse{
		Success: true,
		Message: "Weather Consumer API is healthy",
		Data: map[string]interface{}{
			"timestamp": time.Now(),
			"version":   "1.0.0",
			"service":   "weather-consumer",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getMetrics returns basic metrics information
func (api *WeatherAPI) getMetrics(w http.ResponseWriter, r *http.Request) {
	response := WeatherResponse{
		Success: true,
		Message: "Metrics endpoint available",
		Data: map[string]interface{}{
			"prometheus_endpoint": "/metrics",
			"description":         "Prometheus metrics are available at /metrics endpoint",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// setTestTemperature sets a test temperature for alerting purposes
func (api *WeatherAPI) setTestTemperature(w http.ResponseWriter, r *http.Request) {
	var req struct {
		City        string  `json:"city"`
		Temperature float64 `json:"temperature"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.City == "" {
		api.sendErrorResponse(w, "city is required", http.StatusBadRequest)
		return
	}

	// Set test temperature in metrics
	metrics := api.consumer.GetMetrics()
	metrics.SetTestTemperature(req.City, req.Temperature)

	response := WeatherResponse{
		Success: true,
		Message: fmt.Sprintf("Test temperature set for %s: %.1f°C", req.City, req.Temperature),
		Data: map[string]interface{}{
			"city":        req.City,
			"temperature": req.Temperature,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sendErrorResponse sends an error response
func (api *WeatherAPI) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := WeatherResponse{
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// StartServer starts the HTTP server
func (api *WeatherAPI) StartServer(port string) {
	log.Printf("🌐 Starting Weather Consumer API server on port %s", port)
	log.Printf("📊 Metrics available at http://localhost:%s/metrics", port)
	log.Printf("🔍 Health check at http://localhost:%s/health", port)

	if err := http.ListenAndServe(":"+port, api.router); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
