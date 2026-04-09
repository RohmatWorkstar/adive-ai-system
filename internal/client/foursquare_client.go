package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"map-backend/internal/model"
)

// FoursquareClient handles place search via the Foursquare Places API v3
type FoursquareClient interface {
	SearchPlaces(query string, lat, lng float64, limit int) ([]model.Place, error)
}

type foursquareClientImpl struct {
	apiKey     string
	httpClient *http.Client
}

// NewFoursquareClient creates a new Foursquare client with a reusable HTTP client
func NewFoursquareClient(apiKey string) FoursquareClient {
	return &foursquareClientImpl{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// foursquareResponse represents the raw API response from Foursquare v3
type foursquareResponse struct {
	Results []struct {
		Name     string `json:"name"`
		Geocodes struct {
			Main struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"main"`
		} `json:"geocodes"`
		Location struct {
			FormattedAddress string `json:"formatted_address"`
			Address          string `json:"address"`
			Locality         string `json:"locality"`
			Region           string `json:"region"`
			Country          string `json:"country"`
		} `json:"location"`
		Categories []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"categories"`
		Rating float64 `json:"rating"`
	} `json:"results"`
}

// SearchPlaces queries Foursquare for real place data
// If lat/lng are 0, it uses Jakarta as the default location
func (c *foursquareClientImpl) SearchPlaces(query string, lat, lng float64, limit int) ([]model.Place, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("foursquare: API key not configured")
	}

	// Default to Jakarta if no coordinates provided
	if lat == 0 && lng == 0 {
		lat = -6.2088
		lng = 106.8456
		log.Println("[Foursquare] No coordinates provided, defaulting to Jakarta")
	}

	if limit <= 0 || limit > 50 {
		limit = 10
	}

	baseURL := "https://api.foursquare.com/v3/places/search"

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("foursquare: failed to create request: %w", err)
	}

	// Set authorization header (Foursquare v3 style)
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Accept", "application/json")

	// Set query parameters
	q := req.URL.Query()
	q.Set("query", query)
	q.Set("ll", fmt.Sprintf("%f,%f", lat, lng))
	q.Set("limit", fmt.Sprintf("%d", limit))
	req.URL.RawQuery = q.Encode()

	log.Printf("[Foursquare] Searching: query=%s, ll=%f,%f, limit=%d", query, lat, lng, limit)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("foursquare: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("foursquare: API error status %d", resp.StatusCode)
	}

	var fsResp foursquareResponse
	if err := json.NewDecoder(resp.Body).Decode(&fsResp); err != nil {
		return nil, fmt.Errorf("foursquare: failed to decode response: %w", err)
	}

	// Map Foursquare results to our Place model
	var places []model.Place
	for _, r := range fsResp.Results {
		// Determine address — prefer formatted_address, fallback to components
		address := r.Location.FormattedAddress
		if address == "" {
			address = buildAddress(r.Location.Address, r.Location.Locality, r.Location.Region, r.Location.Country)
		}

		// Determine category — use first category name if available
		category := "General"
		if len(r.Categories) > 0 {
			category = r.Categories[0].Name
		}

		places = append(places, model.Place{
			Name:     r.Name,
			Rating:   r.Rating,
			Address:  address,
			Lat:      r.Geocodes.Main.Latitude,
			Lng:      r.Geocodes.Main.Longitude,
			Category: category,
		})
	}

	log.Printf("[Foursquare] Found %d places", len(places))
	return places, nil
}

// buildAddress constructs a readable address from individual components
func buildAddress(parts ...string) string {
	var result string
	for _, p := range parts {
		if p != "" {
			if result != "" {
				result += ", "
			}
			result += p
		}
	}
	if result == "" {
		return "Address not available"
	}
	return result
}
