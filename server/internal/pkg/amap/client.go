package amap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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

// LatLng is a single geographic point.
type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// DrivingResult is the decoded road route between an origin and destination,
// optionally passing through waypoints. Points are the full polyline to draw.
type DrivingResult struct {
	Points    []LatLng `json:"points"`
	DistanceM int      `json:"distance"`
	DurationS int      `json:"duration"`
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

// fmtCoord formats a coordinate as Amap expects: "lng,lat" with 6 decimals.
func fmtCoord(p LatLng) string {
	return strconv.FormatFloat(p.Lng, 'f', 6, 64) + "," + strconv.FormatFloat(p.Lat, 'f', 6, 64)
}

// Driving requests a road route from origin to dest passing through waypoints
// (Amap allows at most 16 waypoints per request). It returns the full polyline
// to render plus total distance (m) and duration (s).
func (a *Client) Driving(ctx context.Context, origin, dest LatLng, waypoints []LatLng) (*DrivingResult, error) {
	if a.apiKey == "" {
		return nil, fmt.Errorf("amap api key not configured")
	}
	if len(waypoints) > 16 {
		return nil, fmt.Errorf("too many waypoints: %d (max 16)", len(waypoints))
	}
	params := url.Values{}
	params.Set("key", a.apiKey)
	params.Set("origin", fmtCoord(origin))
	params.Set("destination", fmtCoord(dest))
	params.Set("extensions", "base")
	if len(waypoints) > 0 {
		parts := make([]string, len(waypoints))
		for i, w := range waypoints {
			parts[i] = fmtCoord(w)
		}
		params.Set("waypoints", strings.Join(parts, ";"))
	}
	endpoint := "https://restapi.amap.com/v3/direction/driving?" + params.Encode()

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
		Status string `json:"status"`
		Info   string `json:"info"`
		Route  struct {
			Paths []struct {
				Distance string `json:"distance"`
				Duration string `json:"duration"`
				Steps    []struct {
					Polyline string `json:"polyline"`
				} `json:"steps"`
			} `json:"paths"`
		} `json:"route"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Status != "1" || len(payload.Route.Paths) == 0 {
		return nil, fmt.Errorf("driving route failed: %s", payload.Info)
	}
	path := payload.Route.Paths[0]
	var points []LatLng
	for _, step := range path.Steps {
		for _, pair := range strings.Split(step.Polyline, ";") {
			if pair == "" {
				continue
			}
			var lng, lat float64
			if _, err := fmt.Sscanf(pair, "%f,%f", &lng, &lat); err != nil {
				continue
			}
			points = append(points, LatLng{Lat: lat, Lng: lng})
		}
	}
	dist, _ := strconv.Atoi(path.Distance)
	dur, _ := strconv.Atoi(path.Duration)
	return &DrivingResult{Points: points, DistanceM: dist, DurationS: dur}, nil
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
