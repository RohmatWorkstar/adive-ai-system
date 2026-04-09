package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"map-backend/internal/client"
	"map-backend/internal/model"
	"map-backend/internal/repository"
	"map-backend/internal/utils"
)

// SearchService defines the search business logic interface
type SearchService interface {
	Search(ctx context.Context, query string, userID string) (*model.SearchResponse, error)
}

// cacheEntry holds a cached response with a timestamp for TTL
type cacheEntry struct {
	response  *model.SearchResponse
	createdAt time.Time
}

type searchServiceImpl struct {
	repo       repository.SearchRepository
	aiClient   client.AIClient
	fsqClient  client.FoursquareClient
	memCache   map[string]*cacheEntry
	cacheMu    sync.RWMutex
	cacheTTL   time.Duration
}

// NewSearchService creates the search service with in-memory cache
func NewSearchService(
	repo repository.SearchRepository,
	aiClient client.AIClient,
	fsqClient client.FoursquareClient,
) SearchService {
	svc := &searchServiceImpl{
		repo:      repo,
		aiClient:  aiClient,
		fsqClient: fsqClient,
		memCache:  make(map[string]*cacheEntry),
		cacheTTL:  24 * time.Hour,
	}

	// Start background goroutine to evict expired cache entries
	go svc.cacheCleanup()

	return svc
}

func (s *searchServiceImpl) Search(ctx context.Context, query string, userID string) (*model.SearchResponse, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}

	queryKey := utils.NormalizeQuery(query)

	// ── Step 1: Check in-memory cache ──
	if cached := s.getFromMemCache(queryKey); cached != nil {
		log.Printf("[Cache] In-memory cache HIT for: %s", queryKey)
		return cached, nil
	}

	// ── Step 2: Check DB cache ──
	cachedResponse, err := s.repo.GetCache(ctx, queryKey)
	if err == nil && cachedResponse != nil {
		log.Printf("[Cache] DB cache HIT for: %s", queryKey)
		// Promote to in-memory cache
		s.setMemCache(queryKey, cachedResponse)
		return cachedResponse, nil
	}

	// ── Step 3: Extract Intent via AI (Grok → Gemini → keyword fallback) ──
	intent, err := s.aiClient.ExtractIntent(query)
	if err != nil {
		log.Printf("[Search] AI Intent extraction failed: %v. Using raw query.", err)
		intent = &model.AIIntent{
			Query: query,
		}
	}

	// Save search history asynchronously
	go func() {
		bgCtx := context.Background()
		if saveErr := s.repo.SaveHistory(bgCtx, query, userID, intent); saveErr != nil {
			log.Printf("[Search] Failed to save history: %v", saveErr)
		}
	}()

	// ── Step 4: Determine location coordinates ──
	lat, lng := resolveCoordinates(intent.Location)

	// ── Step 5: Build Foursquare search query ──
	searchQuery := buildSearchQuery(intent)

	// ── Step 6: Fetch places from Foursquare ──
	places, err := s.fsqClient.SearchPlaces(searchQuery, lat, lng, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search places: %w", err)
	}

	if len(places) == 0 {
		return &model.SearchResponse{
			Places:  []model.Place{},
			Summary: "I couldn't find any places matching your query. Try a different search term or location.",
		}, nil
	}

	// ── Step 7: Rank places using the intent-aware algorithm ──
	rankedPlaces := utils.RankPlaces(places, intent)

	// Top places for summary (max 5)
	topPlaces := rankedPlaces
	if len(topPlaces) > 5 {
		topPlaces = topPlaces[:5]
	}

	// ── Step 8: Generate summary via AI (Gemini → Grok → static) ──
	// Run in a goroutine with a channel for the result
	type summaryResult struct {
		summary string
		err     error
	}
	summaryCh := make(chan summaryResult, 1)
	go func() {
		s, err := s.aiClient.GenerateSummary(topPlaces)
		summaryCh <- summaryResult{summary: s, err: err}
	}()

	// Wait for summary with timeout
	var summary string
	select {
	case res := <-summaryCh:
		if res.err != nil {
			log.Printf("[Search] AI Summary failed: %v. Using static fallback.", res.err)
			summary = "Here are some great places based on your search!"
		} else {
			summary = res.summary
		}
	case <-time.After(15 * time.Second):
		log.Println("[Search] AI Summary timed out. Using static fallback.")
		summary = "Here are some great places based on your search!"
	}

	response := &model.SearchResponse{
		Places:  rankedPlaces,
		Summary: summary,
	}

	// ── Step 9: Save to caches asynchronously ──
	go func() {
		bgCtx := context.Background()
		if saveErr := s.repo.SaveCache(bgCtx, queryKey, response); saveErr != nil {
			log.Printf("[Search] Failed to save DB cache: %v", saveErr)
		}
	}()
	s.setMemCache(queryKey, response)

	return response, nil
}

// ── In-memory cache methods ──

func (s *searchServiceImpl) getFromMemCache(key string) *model.SearchResponse {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	entry, exists := s.memCache[key]
	if !exists {
		return nil
	}

	// Check TTL
	if time.Since(entry.createdAt) > s.cacheTTL {
		return nil
	}

	// Deep copy to avoid race conditions on the cached data
	data, _ := json.Marshal(entry.response)
	var copy model.SearchResponse
	json.Unmarshal(data, &copy)
	return &copy
}

func (s *searchServiceImpl) setMemCache(key string, response *model.SearchResponse) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.memCache[key] = &cacheEntry{
		response:  response,
		createdAt: time.Now(),
	}
}

func (s *searchServiceImpl) cacheCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.cacheMu.Lock()
		for key, entry := range s.memCache {
			if time.Since(entry.createdAt) > s.cacheTTL {
				delete(s.memCache, key)
			}
		}
		s.cacheMu.Unlock()
	}
}

// ── Helper functions ──

// resolveCoordinates converts a location name to lat/lng using known locations
// Falls back to Jakarta (-6.2088, 106.8456) if unknown
func resolveCoordinates(location string) (float64, float64) {
	if location == "" {
		return 0, 0 // Foursquare client will default to Jakarta
	}

	loc := strings.ToLower(strings.TrimSpace(location))

	// Known Indonesian cities with their coordinates
	knownLocations := map[string][2]float64{
		"jakarta":     {-6.2088, 106.8456},
		"bekasi":      {-6.2383, 106.9756},
		"bandung":     {-6.9175, 107.6191},
		"surabaya":    {-7.2575, 112.7521},
		"yogyakarta":  {-7.7972, 110.3688},
		"semarang":    {-6.9666, 110.4196},
		"medan":       {3.5952, 98.6722},
		"makassar":    {-5.1477, 119.4327},
		"palembang":   {-2.9761, 104.7754},
		"tangerang":   {-6.1702, 106.6403},
		"depok":       {-6.4025, 106.7942},
		"bogor":       {-6.5971, 106.8060},
		"malang":      {-7.9666, 112.6326},
		"solo":        {-7.5755, 110.8243},
		"denpasar":    {-8.6705, 115.2126},
		"bali":        {-8.3405, 115.0920},
	}

	if coords, ok := knownLocations[loc]; ok {
		return coords[0], coords[1]
	}

	return 0, 0 // Unknown location — let Foursquare client default to Jakarta
}

// buildSearchQuery constructs a Foursquare search query string from the AI intent
func buildSearchQuery(intent *model.AIIntent) string {
	parts := []string{}

	if intent.Query != "" {
		parts = append(parts, intent.Query)
	}

	if intent.Category != "" && intent.Category != intent.Query {
		parts = append(parts, intent.Category)
	}

	if len(parts) == 0 {
		return "restaurant" // sensible default
	}

	return strings.Join(parts, " ")
}
