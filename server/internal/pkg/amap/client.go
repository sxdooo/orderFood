package amap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"

	"github.com/orderfood/server/internal/config"
)

type Client struct {
	apiKey string
	client *http.Client
}

type GeocodeResult struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func NewClient(cfg *config.Config) *Client {
	return &Client{apiKey: cfg.AmapAPIKey, client: http.DefaultClient}
}

func (a *Client) Geocode(ctx context.Context, address string) (*GeocodeResult, error) {
	if a.apiKey == "" {
		return nil, fmt.Errorf("amap api key not configured")
	}
	endpoint := fmt.Sprintf(
		"https://restapi.amap.com/v3/geocode/geo?key=%s&address=%s",
		url.QueryEscape(a.apiKey),
		url.QueryEscape(address),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Status   string `json:"status"`
		Geocodes []struct {
			Location string `json:"location"`
		} `json:"geocodes"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Status != "1" || len(payload.Geocodes) == 0 {
		return nil, fmt.Errorf("geocode failed")
	}
	var lat, lng float64
	if _, err := fmt.Sscanf(payload.Geocodes[0].Location, "%f,%f", &lng, &lat); err != nil {
		return nil, err
	}
	return &GeocodeResult{Lat: lat, Lng: lng}, nil
}

func HaversineDistanceKm(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371.0
	rad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := rad(lat2 - lat1)
	dLng := rad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(lat1))*math.Cos(rad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
