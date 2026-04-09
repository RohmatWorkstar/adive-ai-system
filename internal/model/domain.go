package model

import "time"

// SearchRequest is the incoming request body for POST /api/search
type SearchRequest struct {
	Query string `json:"query"`
}

// AIIntent represents the structured intent extracted by AI from a natural language query
type AIIntent struct {
	Query    string   `json:"query"`
	Location string   `json:"location"`
	Category string   `json:"category,omitempty"`
	Filters  []string `json:"filters"`
}

// Place represents a single place result from Foursquare
type Place struct {
	Name     string  `json:"name"`
	Rating   float64 `json:"rating"`
	Address  string  `json:"address"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Category string  `json:"category"`
	Score    float64 `json:"-"` // Internal ranking score, hidden from JSON output
}

// SearchResponse is the final response returned to the frontend
type SearchResponse struct {
	Places  []Place `json:"places"`
	Summary string  `json:"summary"`
}

// Favorite represents a user's saved favorite place
type Favorite struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PlaceName string    `json:"place_name"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
}

// FavoriteRequest is the incoming request body for adding a favorite
type FavoriteRequest struct {
	PlaceName string  `json:"place_name"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Address   string  `json:"address"`
}
