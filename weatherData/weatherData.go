package weatherdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"heat-transfer/interop"
	"io"
	"net/http"
	"strings"
	"time"
)

type HourlyForecast struct {
	Dt   int64   `json:"dt"`
	Temp float64 `json:"temp"`
}

type ForecastData struct {
	Hourly         []HourlyForecast `json:"hourly"`
	TimezoneOffset int              `json:"timezone_offset"`
}

type GeoLocation struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Country string  `json:"country"`
}

func GetCityTemperatureForecastNow(query, apiKey string) ([840]float64, error) {
	query = strings.ReplaceAll(query, " ", "%20")

	geoURL := fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", query, apiKey)
	fmt.Println("Geocoding URL:", geoURL)
	resp, err := http.Get(geoURL)
	if err != nil {
		return [840]float64{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return [840]float64{}, fmt.Errorf("geocoding API error: %s", resp.Status)
	}
	geoBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return [840]float64{}, err
	}
	var geoResults []GeoLocation
	if err := json.Unmarshal(geoBody, &geoResults); err != nil {
		return [840]float64{}, err
	}
	if len(geoResults) == 0 {
		return [840]float64{}, errors.New("no geocoding results found")
	}
	lat, lon := geoResults[0].Lat, geoResults[0].Lon
	fmt.Println("Coordinates:", lat, lon)

	forecastURL := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%f&lon=%f&units=metric&appid=%s", lat, lon, apiKey)
	fmt.Printf("Forecast URL: %s\n", forecastURL)
	resp, err = http.Get(forecastURL)
	if err != nil {
		return [840]float64{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return [840]float64{}, fmt.Errorf("forecast API error: %s", resp.Status)
	}
	forecastBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return [840]float64{}, err
	}
	var forecast ForecastData
	if err := json.Unmarshal(forecastBody, &forecast); err != nil {
		return [840]float64{}, err
	}

	loc := time.FixedZone("local", forecast.TimezoneOffset)

	now := time.Now().In(loc)
	targetDate := now
	if now.Hour() >= 7 {
		targetDate = now.Add(24 * time.Hour)
	}
	targetDayStart := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 5, 0, 0, 0, loc)
	targetDayEnd := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 19, 0, 0, 0, loc)

	hourlyTemps := make([]float64, 0)

	for _, hourly := range forecast.Hourly {
		hourlyTime := time.Unix(hourly.Dt, 0).In(loc)
		if hourlyTime.Before(targetDayStart) || hourlyTime.After(targetDayEnd) {
			continue
		}

		hourlyTemps = append(hourlyTemps, hourly.Temp)
	}

	// interpolate

	return interop.MovingWindowInterpolateTemperature(hourlyTemps), nil
}

// doesnt work
func GetCityTemperatureForecastHistorical(query, apiKey string, t time.Time) ([840]float64, error) {
	query = strings.ReplaceAll(query, " ", "%20")

	geoURL := fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", query, apiKey)
	fmt.Println("Geocoding URL:", geoURL)
	resp, err := http.Get(geoURL)
	if err != nil {
		return [840]float64{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return [840]float64{}, fmt.Errorf("geocoding API error: %s", resp.Status)
	}
	geoBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return [840]float64{}, err
	}
	var geoResults []GeoLocation
	if err := json.Unmarshal(geoBody, &geoResults); err != nil {
		return [840]float64{}, err
	}
	if len(geoResults) == 0 {
		return [840]float64{}, errors.New("no geocoding results found")
	}
	lat, lon := geoResults[0].Lat, geoResults[0].Lon
	fmt.Println("Coordinates:", lat, lon)

	// doesnt work
	historicalURL := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall/timemachine?lat=%f&lon=%f&dt=%d&units=metric&appid=%s", lat, lon, t.Unix(), apiKey)
	fmt.Printf("Historical URL: %s\n", historicalURL)
	resp, err = http.Get(historicalURL)
	if err != nil {
		return [840]float64{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return [840]float64{}, fmt.Errorf("historical API error: %s", resp.Status)
	}
	historicalBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return [840]float64{}, err
	}
	var forecast ForecastData
	if err := json.Unmarshal(historicalBody, &forecast); err != nil {
		return [840]float64{}, err
	}

	// local time
	loc := time.FixedZone("local", forecast.TimezoneOffset)

	targetDate := t.In(loc)
	targetDayStart := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 5, 0, 0, 0, loc)
	targetDayEnd := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 19, 0, 0, 0, loc)

	hourlyTemps := make([]float64, 0)

	for _, hourly := range forecast.Hourly {
		// only get the data from 5 AM to 7 PM.
		hourlyTime := time.Unix(hourly.Dt, 0).In(loc)
		if hourlyTime.Before(targetDayStart) || hourlyTime.After(targetDayEnd) {
			continue
		}

		hourlyTemps = append(hourlyTemps, hourly.Temp)
	}

	// interpolate
	return interop.MovingWindowInterpolateTemperature(hourlyTemps), nil
}