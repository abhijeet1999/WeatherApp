package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/abhijeet1999/weather/models"
)

// parseLine parses a single line in format "zipcode,days,temp_threshold,wind_threshold,humidity_threshold"
func parseLine(line string) (models.WeatherRequest, error) {
	parts := strings.Split(line, ",")

	// Check for correct number of fields
	if len(parts) != 5 {
		if len(parts) < 5 {
			return models.WeatherRequest{}, fmt.Errorf("invalid format: expected 5 fields (zipcode,days,temp,wind,humidity), got %d fields in '%s'", len(parts), line)
		} else {
			return models.WeatherRequest{}, fmt.Errorf("invalid format: expected 5 fields (zipcode,days,temp,wind,humidity), got %d fields in '%s' (extra fields detected)", len(parts), line)
		}
	}

	// Validate and parse zip code
	zipCode := strings.TrimSpace(parts[0])
	if err := validateZipCode(zipCode); err != nil {
		return models.WeatherRequest{}, fmt.Errorf("invalid zip code: %v", err)
	}

	// Validate and parse days
	daysStr := strings.TrimSpace(parts[1])
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		return models.WeatherRequest{}, fmt.Errorf("invalid days value '%s': must be a number", daysStr)
	}
	if days < 1 || days > 5 {
		return models.WeatherRequest{}, fmt.Errorf("invalid days value %d: must be between 1 and 5", days)
	}

	// Validate and parse temperature threshold
	tempStr := strings.TrimSpace(parts[2])
	alertTemp, err := strconv.ParseFloat(tempStr, 32)
	if err != nil {
		return models.WeatherRequest{}, fmt.Errorf("invalid temperature threshold '%s': must be a number", tempStr)
	}
	if alertTemp < -50 || alertTemp > 60 {
		return models.WeatherRequest{}, fmt.Errorf("invalid temperature threshold %.1f: must be between -50°C and 60°C", alertTemp)
	}

	// Validate and parse wind threshold
	windStr := strings.TrimSpace(parts[3])
	alertWind, err := strconv.ParseFloat(windStr, 32)
	if err != nil {
		return models.WeatherRequest{}, fmt.Errorf("invalid wind threshold '%s': must be a number", windStr)
	}
	if alertWind < 0 || alertWind > 100 {
		return models.WeatherRequest{}, fmt.Errorf("invalid wind threshold %.1f: must be between 0 and 100 m/s", alertWind)
	}

	// Validate and parse humidity threshold
	humidityStr := strings.TrimSpace(parts[4])
	alertHumidity, err := strconv.Atoi(humidityStr)
	if err != nil {
		return models.WeatherRequest{}, fmt.Errorf("invalid humidity threshold '%s': must be a number", humidityStr)
	}
	if alertHumidity < 0 || alertHumidity > 100 {
		return models.WeatherRequest{}, fmt.Errorf("invalid humidity threshold %d: must be between 0 and 100%%", alertHumidity)
	}

	return models.WeatherRequest{
		ZipCode:       zipCode,
		Days:          days,
		AlertTemp:     float32(alertTemp),
		AlertWind:     float32(alertWind),
		AlertHumidity: alertHumidity,
	}, nil
}

// validateZipCode validates a US zip code format
func validateZipCode(zipCode string) error {
	if zipCode == "" {
		return fmt.Errorf("zip code cannot be empty")
	}

	// Check if it's a valid US zip code format (5 digits or 5+4 format)
	if len(zipCode) == 5 {
		// Check if all characters are digits
		for _, char := range zipCode {
			if char < '0' || char > '9' {
				return fmt.Errorf("'%s' contains non-numeric characters", zipCode)
			}
		}
		return nil
	} else if len(zipCode) == 10 && zipCode[5] == '-' {
		// Check 5+4 format (12345-6789)
		for i, char := range zipCode {
			if i == 5 {
				continue // Skip the dash
			}
			if char < '0' || char > '9' {
				return fmt.Errorf("'%s' contains non-numeric characters", zipCode)
			}
		}
		return nil
	} else {
		return fmt.Errorf("'%s' is not a valid US zip code format (expected 5 digits or 5+4 format)", zipCode)
	}
}

// ParseInputFile reads and parses the input.txt file
func ParseInputFile(filename string) ([]models.WeatherRequest, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	var requests []models.WeatherRequest
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		request, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line %d '%s': %w", lineNumber, line, err)
		}

		requests = append(requests, request)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if len(requests) == 0 {
		return nil, fmt.Errorf("no valid entries found in input file")
	}

	return requests, nil
}
