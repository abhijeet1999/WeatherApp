package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/abhijeet1999/weather/models"
)

// ParseInputFile reads and parses the input.txt file
func ParseInputFile(filename string) ([]models.WeatherRequest, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	var requests []models.WeatherRequest
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		request, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line '%s': %w", line, err)
		}

		requests = append(requests, request)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return requests, nil
}

// parseLine parses a single line in format "zipcode,days,alert_temp"
func parseLine(line string) (models.WeatherRequest, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 3 {
		return models.WeatherRequest{}, fmt.Errorf("invalid format: expected 'zipcode,days,alert_temp', got '%s'", line)
	}

	zipCode := strings.TrimSpace(parts[0])
	if zipCode == "" {
		return models.WeatherRequest{}, fmt.Errorf("zipcode cannot be empty")
	}

	days, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return models.WeatherRequest{}, fmt.Errorf("invalid days value: %s", parts[1])
	}

	if days < 1 || days > 5 {
		return models.WeatherRequest{}, fmt.Errorf("days must be between 1 and 5, got %d", days)
	}

	alertTemp, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 32)
	if err != nil {
		return models.WeatherRequest{}, fmt.Errorf("invalid alert temperature: %s", parts[2])
	}

	return models.WeatherRequest{
		ZipCode:   zipCode,
		Days:      days,
		AlertTemp: float32(alertTemp),
	}, nil
}
