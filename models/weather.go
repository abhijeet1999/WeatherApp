package models

// OpenWeatherResponse represents the response from OpenWeatherMap Current Weather API
type OpenWeatherResponse struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []OpenWeatherCondition `json:"weather"`
	Base    string                 `json:"base"`
	Main    struct {
		Temp      float32 `json:"temp"`
		FeelsLike float32 `json:"feels_like"`
		TempMin   float32 `json:"temp_min"`
		TempMax   float32 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float32 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt  int64 `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		Id      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int64  `json:"sunrise"`
		Sunset  int64  `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

// OpenWeatherForecastResponse represents the response from OpenWeatherMap 5-Day Forecast API
type OpenWeatherForecastResponse struct {
	Cod     string         `json:"cod"`
	Message int            `json:"message"`
	Cnt     int            `json:"cnt"`
	List    []ForecastItem `json:"list"`
	City    struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Coord struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"coord"`
		Country    string `json:"country"`
		Population int    `json:"population"`
		Timezone   int    `json:"timezone"`
		Sunrise    int64  `json:"sunrise"`
		Sunset     int64  `json:"sunset"`
	} `json:"city"`
}

// ForecastItem represents a single forecast entry
type ForecastItem struct {
	Dt   int64 `json:"dt"`
	Main struct {
		Temp      float32 `json:"temp"`
		FeelsLike float32 `json:"feels_like"`
		TempMin   float32 `json:"temp_min"`
		TempMax   float32 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []OpenWeatherCondition `json:"weather"`
	Clouds  struct {
		All int `json:"all"`
	} `json:"clouds"`
	Wind struct {
		Speed float32 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Visibility int     `json:"visibility"`
	Pop        float32 `json:"pop"`
	Sys        struct {
		Pod string `json:"pod"`
	} `json:"sys"`
	DtTxt string `json:"dt_txt"`
}

// OpenWeatherCondition represents weather condition details
type OpenWeatherCondition struct {
	Id          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// GeoResponse represents the response from OpenWeatherMap Geocoding API
type GeoResponse struct {
	Zip     string  `json:"zip"`
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Country string  `json:"country"`
}

// WeatherRequest represents the input parameters for weather requests
type WeatherRequest struct {
	ZipCode   string
	Days      int
	AlertTemp float32
}

// DailyForecast represents a daily weather summary
type DailyForecast struct {
	Date        string
	TempMin     float32
	TempMax     float32
	Description string
	Humidity    int
	WindSpeed   float32
	Icon        string
	HasAlert    bool
}
