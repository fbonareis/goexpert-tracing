package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const baseURLviaCEP = "http://viacep.com.br"

type LocationResponse struct {
	City string `json:"localidade"`
	Erro bool   `json:"erro"`
}

func GetLocation(ctx context.Context, zipCode string) (*LocationResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/ws/%s/json", baseURLviaCEP, zipCode), nil)
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
	var l LocationResponse
	if err = json.Unmarshal(body, &l); err != nil {
		return nil, err
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	return &l, nil
}
