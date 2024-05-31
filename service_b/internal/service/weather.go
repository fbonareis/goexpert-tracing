package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const baseURLweatherAPI = "http://api.weatherapi.com/v1"

type WeatherResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
		TempF float64 `json:"temp_f"`
	} `json:"current"`
}

func (w *WeatherResponse) GetTempF() float64 {
	return roundFloat(w.Current.TempC*1.8+32, 2)
}
func (w *WeatherResponse) GetTempK() float64 {
	return roundFloat(w.Current.TempC+273, 2)
}

func GetWeatherFromCity(ctx context.Context, city string) (*WeatherResponse, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return nil, errors.New("weather api key not found")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/current.json?key=%s&q=%s&aqi=no", baseURLweatherAPI, apiKey, normalize(city)), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var w WeatherResponse
	if err = json.Unmarshal(body, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

func normalize(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	output, _, e := transform.String(t, s)
	if e != nil {
		panic(e)
	}
	return strings.ReplaceAll(output, " ", "%20")
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
