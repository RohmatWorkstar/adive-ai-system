package model

import "time"

type SearchRequest struct {
	Query string `json:"query"`
}

type AIIntent struct {
	Location string   `json:"location"`
	Category string   `json:"category"`
	Filters  []string `json:"filters"`
}

type Place struct {
	Name       string  `json:"name"`
	Rating     float64 `json:"rating"`
	Address    string  `json:"address"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	PriceLevel int     `json:"price_level"`
	Score      float64 `json:"score,omitempty"` // For internal ranking
}

type SearchResponse struct {
	Places  []Place `json:"places"`
	Summary string  `json:"summary"`
}

type Favorite struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PlaceName string    `json:"place_name"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
}

type FavoriteRequest struct {
	PlaceName string  `json:"place_name"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Address   string  `json:"address"`
}
