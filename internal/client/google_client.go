package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"map-backend/internal/model"
)

type GoogleClient interface {
	SearchPlaces(query string) ([]model.Place, error)
}

type googleClientImpl struct {
	apiKey     string
	httpClient *http.Client
}

func NewGoogleClient(apiKey string) GoogleClient {
	return &googleClientImpl{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *googleClientImpl) SearchPlaces(query string) ([]model.Place, error) {
	baseURL := "https://maps.googleapis.com/maps/api/place/textsearch/json"
	
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	q := reqURL.Query()
	q.Add("query", query)
	q.Add("key", c.apiKey)
	reqURL.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(reqURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google places api error: returned status %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			Name             string  `json:"name"`
			FormattedAddress string  `json:"formatted_address"`
			Rating           float64 `json:"rating"`
			PriceLevel       int     `json:"price_level"`
			Geometry         struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"results"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "OK" && result.Status != "ZERO_RESULTS" {
		return nil, fmt.Errorf("google places api error status: %s", result.Status)
	}

	var places []model.Place
	for _, r := range result.Results {
		places = append(places, model.Place{
			Name:       r.Name,
			Address:    r.FormattedAddress,
			Rating:     r.Rating,
			PriceLevel: r.PriceLevel,
			Lat:        r.Geometry.Location.Lat,
			Lng:        r.Geometry.Location.Lng,
		})
	}

	return places, nil
}
