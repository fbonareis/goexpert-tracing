package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/fbonareis/goexpert-tracing-service-b/internal/service"
)

var (
	ErrCanNotFindZipCode = errors.New("can not find zipcode")
)

type Webserver struct {
	Tracer trace.Tracer
}

func NewServer(tracer trace.Tracer) *Webserver {
	return &Webserver{
		Tracer: tracer,
	}
}

type LocationWeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func (we *Webserver) CreateServer() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Get("/{cep}", we.HandleRequest)
	return router
}

func (h *Webserver) HandleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	ctx, span := h.Tracer.Start(ctx, "location-weather-request")
	defer span.End()

	cep := chi.URLParam(r, "cep")

	ctx, spanLocation := h.Tracer.Start(ctx, "external request - search-location-on-viacep")
	location, err := service.GetLocation(ctx, cep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if location.Erro {
		http.Error(w, ErrCanNotFindZipCode.Error(), http.StatusNotFound)
		return
	}
	spanLocation.End()

	ctx, spanWeather := h.Tracer.Start(ctx, "external request - get-city-weather-on-weatherapi")
	weather, err := service.GetWeatherFromCity(ctx, location.City)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	spanWeather.End()

	response := &LocationWeatherResponse{
		City:  location.City,
		TempC: weather.Current.TempC,
		TempF: weather.GetTempF(),
		TempK: weather.GetTempK(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
