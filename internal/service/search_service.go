package service

import (
	"context"
	"errors"
	"fmt"

	"log"

	"map-backend/internal/client"
	"map-backend/internal/model"
	"map-backend/internal/repository"
	"map-backend/internal/utils"
)

type SearchService interface {
	Search(ctx context.Context, query string) (*model.SearchResponse, error)
}

type searchServiceImpl struct {
	repo       repository.SearchRepository
	aiClient   client.AIClient
	googleMaps client.GoogleClient
}

func NewSearchService(repo repository.SearchRepository, aiClient client.AIClient, googleMaps client.GoogleClient) SearchService {
	return &searchServiceImpl{
		repo:       repo,
		aiClient:   aiClient,
		googleMaps: googleMaps,
	}
}

func (s *searchServiceImpl) Search(ctx context.Context, query string) (*model.SearchResponse, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}

	queryKey := utils.NormalizeQuery(query)

	// 1. Check Cache
	cachedResponse, err := s.repo.GetCache(ctx, queryKey)
	if err == nil && cachedResponse != nil {
		return cachedResponse, nil
	}

	// 2. Extract Intent via AI
	intent, err := s.aiClient.ExtractIntent(query)
	if err != nil {
		log.Printf("AI Intent Extraction skipped or failed: %v. Falling back to raw query.", err)
		// Fallback intent: use the raw query as category
		intent = &model.AIIntent{
			Category: query,
		}
	}

	// Save history asynchronously (or synchronously for simplicity here)
	_ = s.repo.SaveHistory(ctx, query, intent)

	// 3. Search via Google Maps
	// We form a targeted query based on intent
	searchQuery := intent.Category
	if intent.Location != "" {
		searchQuery += " in " + intent.Location
	}
	if len(intent.Filters) > 0 {
		for _, f := range intent.Filters {
			searchQuery += " " + f
		}
	}

	// Fallback to original query if intent extraction is too sparse
	if searchQuery == "" {
		searchQuery = query
	}

	places, err := s.googleMaps.SearchPlaces(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search places: %w", err)
	}

	if len(places) == 0 {
		return &model.SearchResponse{
			Places:  []model.Place{},
			Summary: "I couldn't find any places matching your query.",
		}, nil
	}

	// 4. Rank Places
	rankedPlaces := utils.RankPlaces(places)

	// Optional: Limit to top 5 for summary
	topPlaces := rankedPlaces
	if len(topPlaces) > 5 {
		topPlaces = topPlaces[:5]
	}

	// 5. Generate Summary via AI
	summary, err := s.aiClient.GenerateSummary(topPlaces)
	if err != nil {
		log.Printf("AI Summary generation skipped or failed: %v. Using default summary.", err)
		// Non-fatal error, we can still return places
		summary = "Here are some great places based on your search! (AI summary unavailable)"
	}

	response := &model.SearchResponse{
		Places:  rankedPlaces,
		Summary: summary,
	}

	// 6. Save to Cache
	_ = s.repo.SaveCache(ctx, queryKey, response)

	return response, nil
}
