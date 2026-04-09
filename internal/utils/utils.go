package utils

import (
	"sort"
	"strings"

	"map-backend/internal/model"
)

func NormalizeQuery(query string) string {
	q := strings.TrimSpace(query)
	q = strings.ToLower(q)
	// Additional normalization like removing special characters could be added
	return q
}

func RankPlaces(places []model.Place) []model.Place {
	// Simple ranking algorithm: Score = (Rating * 2) - PriceLevel
	// High rating is good, high price is penalized (just an example of custom ranking)
	for i := range places {
		score := (places[i].Rating * 2.0) - float64(places[i].PriceLevel)
		places[i].Score = score
	}

	// Sort places by Score descending
	sort.Slice(places, func(i, j int) bool {
		return places[i].Score > places[j].Score
	})

	return places
}
