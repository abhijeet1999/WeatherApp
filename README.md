# WeatherApp - Real-time Weather Data Pipeline

A complete weather data pipeline built with Go, Kafka, Prometheus, and Grafana that fetches weather data, processes it through Kafka, and provides real-time monitoring with alerting capabilities.

## ⚡ One-Liner Setup

```bash
export WEATHER_API_KEY="your_key" && docker-compose up --build -d
```

## 🏗️ Architecture

```
[Weather API] → [Go Producer] → [Kafka] → [Go Consumer] → [Prometheus] → [Grafana]
                                                      ↓
                                              [Alert Manager]
```

## 📋 Prerequisites

- Docker and Docker Compose installed
- OpenWeatherMap API key from [https://openweathermap.org/api](https://openweathermap.org/api)

## 🚀 Quick Start

### Simple Setup (No Scripts Required)

```bash
# 1. Set your OpenWeatherMap API key
export WEATHER_API_KEY="your_openweathermap_api_key_here"

# 2. Run the complete system
docker-compose up --build -d

# 3. Access your services
# Grafana Dashboard: http://localhost:3000 (admin/admin)
# Alert Manager: http://localhost:9093/#/alerts
# Prometheus: http://localhost:9090
```

That's it! The system will automatically:
- Build Docker images
- Start Kafka, Prometheus, Grafana, and Alert Manager
- Process weather data for Poughkeepsie, Hoboken, and Beverly Hills
- Generate alerts when Poughkeepsie temperature ≥ 10°C

### Using Pre-built Images (Even Faster)

If you have pre-built images, you can skip the build step:

```bash
# Set your API key
export WEATHER_API_KEY="your_openweathermap_api_key_here"

# Run without building (uses existing images)
docker-compose up -d
```

## 📁 Project Structure

```
WeatherApp/
├── Producer/
│   ├── Dockerfile
│   ├── main.go
│   ├── kafka/
│   │   └── producer.go
│   ├── weather/
│   │   └── service.go
│   └── utils/
│       └── constants.go
├── Consumer/
│   ├── Dockerfile
│   ├── main.go
│   ├── kafka/
│   │   └── consumer.go
│   ├── api/
│   │   └── server.go
│   └── prometheus/
│       └── metrics.go
├── models/
│   ├── weather.go
│   └── geo.go
├── docker-compose.yml
├── prometheus.yml
├── weather_alerts.yml
├── alertmanager.yml
├── input.txt
├── go.mod
├── .gitignore
└── .dockerignore
```

## ⚙️ Configuration

### Environment Variables

- `WEATHER_API_KEY`: Required. Your OpenWeatherMap API key
- `KAFKA_SERVERS`: Kafka broker address (default: kafka:29092)
- `KAFKA_TOPIC`: Kafka topic name (default: weather_data)
- `CONSUMER_GROUP_ID`: Consumer group ID (default: weather-consumer-group)
- `METRICS_PORT`: Prometheus metrics port (default: 8080)
- `API_PORT`: HTTP API port (default: 8081)

### Input Configuration

Edit `input.txt` to specify weather locations:

```
12601,4,10
10001,3,15
90210,2,20
```

**Format**: `zip_code,days_back,interval_hours`
- `zip_code`: US ZIP code for weather location
- `days_back`: Number of days to fetch historical data
- `interval_hours`: Hours between data points

### Alert Configuration

The system includes pre-configured alerts in `weather_alerts.yml`:

- **High Temperature**: > 30°C (warning), > 35°C (critical)
- **Low Temperature**: < 0°C (warning), < -10°C (critical)
- **Poughkeepsie Specific**: ≥ 10°C (warning)
- **High Humidity**: > 90%
- **High Wind Speed**: > 20 m/s

## 🔧 Advanced Usage

### Running Individual Services

```bash
# Start only infrastructure services
docker-compose up -d kafka zookeeper prometheus grafana alertmanager

# Start only weather services  
docker-compose up -d weather-producer weather-consumer

# View logs
docker-compose logs weather-consumer
docker-compose logs weather-producer
```

### Testing Alerts

```bash
# Set test temperature for alerting
curl -X POST http://localhost:8081/test/temperature \
  -H "Content-Type: application/json" \
  -d '{"city": "Poughkeepsie", "temperature": 35.0}'

# Check active alerts
curl http://localhost:9093/api/v2/alerts
```

### Monitoring Commands

```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# View current metrics
curl http://localhost:8080/metrics | grep weather_temperature

# Check Kafka topic
docker-compose exec kafka kafka-topics --bootstrap-server localhost:9092 --list
```

## 🐳 Docker Commands

### Build and Run

```bash
# Build all images
docker-compose build

# Run with fresh build
docker-compose up --build -d

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down --volumes --remove-orphans
```

### Individual Service Management

```bash
# Restart specific service
docker-compose restart weather-consumer

# View service logs
docker-compose logs -f weather-producer

# Execute commands in running container
docker-compose exec weather-consumer sh
```

## 📊 Monitoring and Alerting

### Grafana Dashboards

The system automatically provisions Grafana with:
- Weather temperature trends
- Humidity and wind speed monitoring
- Alert status overview
- System health metrics

### Prometheus Metrics

Available metrics include:
- `weather_temperature_celsius`: Current temperature by city
- `weather_humidity_percent`: Humidity levels
- `weather_wind_speed_ms`: Wind speed
- `weather_pressure_hpa`: Atmospheric pressure

### Alert Manager

Access the Alert Manager UI at http://localhost:9093 to:
- View active alerts
- Silence alerts
- Configure notification channels
- View alert history

## 🔍 Troubleshooting

### Common Issues

1. **Kafka Connection Issues**
   ```bash
   # Check Kafka is running
   docker-compose logs kafka
   
   # Verify network connectivity
   docker-compose exec weather-consumer ping kafka
   ```

2. **API Key Issues**
   ```bash
   # Verify API key is set
   docker-compose exec weather-producer env | grep WEATHER_API_KEY
   ```

3. **Consumer Not Processing Data**
   ```bash
   # Check consumer logs
   docker-compose logs weather-consumer
   
   # Verify Kafka topic has data
   docker-compose exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic weather_data --from-beginning
   ```

### Log Locations

- **Producer**: `docker-compose logs weather-producer`
- **Consumer**: `docker-compose logs weather-consumer`
- **Kafka**: `docker-compose logs kafka`
- **Prometheus**: `docker-compose logs prometheus`
- **Grafana**: `docker-compose logs grafana`

## 🛠️ Development

### Local Development

```bash
# Install Go dependencies
go mod tidy

# Run tests
go test ./...

# Build locally
go build -o producer ./Producer/main.go
go build -o consumer ./Consumer/main.go
```

### Adding New Features

1. **New Metrics**: Add to `Consumer/prometheus/metrics.go`
2. **New Alerts**: Update `weather_alerts.yml`
3. **New API Endpoints**: Modify `Consumer/api/server.go`
4. **New Weather Data**: Extend `models/weather.go`

## 📝 License

This project is for educational and demonstration purposes.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📞 Support

For issues and questions:
1. Check the troubleshooting section
2. Review Docker logs
3. Verify configuration
4. Check service connectivity