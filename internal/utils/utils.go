package utils

import (
	"sort"
	"strings"

	"map-backend/internal/model"
)

// NormalizeQuery normalizes a query string for use as a cache key
func NormalizeQuery(query string) string {
	q := strings.TrimSpace(query)
	q = strings.ToLower(q)
	// Collapse multiple spaces
	fields := strings.Fields(q)
	return strings.Join(fields, " ")
}

// RankPlaces scores and sorts places by relevance
// Scoring: base rating + category match bonus + keyword match bonus + cheap heuristic
func RankPlaces(places []model.Place, intent *model.AIIntent) []model.Place {
	if intent == nil {
		intent = &model.AIIntent{}
	}

	queryLower := strings.ToLower(intent.Query)
	filterSet := make(map[string]bool)
	for _, f := range intent.Filters {
		filterSet[strings.ToLower(f)] = true
	}

	for i := range places {
		score := 0.0

		// Base score from rating (0-10 scale, Foursquare)
		score += places[i].Rating

		// Category match boost (+3)
		if queryLower != "" && strings.Contains(strings.ToLower(places[i].Category), queryLower) {
			score += 3.0
		}

		// Name keyword match boost (+2)
		if queryLower != "" && strings.Contains(strings.ToLower(places[i].Name), queryLower) {
			score += 2.0
		}

		// "cheap" filter heuristic: boost places that might be affordable
		// (simple heuristic based on name keywords)
		if filterSet["cheap"] || filterSet["murah"] {
			nameLower := strings.ToLower(places[i].Name)
			cheapIndicators := []string{"warung", "warteg", "kaki lima", "street", "murah", "sederhana"}
			for _, indicator := range cheapIndicators {
				if strings.Contains(nameLower, indicator) {
					score += 2.0
					break
				}
			}
		}

		places[i].Score = score
	}

	// Sort places by score descending
	sort.Slice(places, func(i, j int) bool {
		return places[i].Score > places[j].Score
	})

	return places
}
