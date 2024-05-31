package web

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Webserver struct {
	Tracer trace.Tracer
}

func NewServer(tracer trace.Tracer) *Webserver {
	return &Webserver{
		Tracer: tracer,
	}
}

func (we *Webserver) CreateServer() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Post("/", we.HandleRequest)
	return router
}

type ZipCodeRequestData struct {
	CEP string `json:"cep"`
}

func (h *Webserver) HandleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))

	ctx, span := h.Tracer.Start(ctx, "get weather information")
	defer span.End()

	var requestData ZipCodeRequestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !isValidCEP(requestData.CEP) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode("invalid zipcode")
		return
	}

	urlServiceB := os.Getenv("SERVICE_B_URL")

	req, err := http.NewRequestWithContext(ctx, "GET", urlServiceB+"/"+requestData.CEP, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(string(bodyBytes))
}

func isValidCEP(cep string) bool {
	matched, _ := regexp.MatchString(`^\d{8}$`, cep)
	return matched
}
