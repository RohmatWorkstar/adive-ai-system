package client

import (
	"log"
	"strings"

	"map-backend/internal/model"
)

// AIClient is the orchestrator that delegates to Grok (primary) and Gemini (fallback)
// for intent extraction, and Gemini (primary) and Grok (fallback) for summary generation
type AIClient interface {
	ExtractIntent(query string) (*model.AIIntent, error)
	GenerateSummary(places []model.Place) (string, error)
}

type aiClientImpl struct {
	grok   *GrokClient
	gemini *GeminiClient
}

// NewAIClient creates a new AI orchestrator with Grok and Gemini providers
func NewAIClient(grok *GrokClient, gemini *GeminiClient) AIClient {
	return &aiClientImpl{
		grok:   grok,
		gemini: gemini,
	}
}

// ExtractIntent tries Grok first, then Gemini, then falls back to simple keyword extraction
func (c *aiClientImpl) ExtractIntent(query string) (*model.AIIntent, error) {
	// Step 1: Try Grok (primary)
	if c.grok.IsAvailable() {
		intent, err := c.grok.ExtractIntent(query)
		if err == nil {
			return intent, nil
		}
		log.Printf("[AI Orchestrator] Grok intent extraction failed: %v. Falling back to Gemini...", err)
	} else {
		log.Println("[AI Orchestrator] Grok not available, trying Gemini...")
	}

	// Step 2: Try Gemini (fallback)
	if c.gemini.IsAvailable() {
		intent, err := c.gemini.ExtractIntent(query)
		if err == nil {
			return intent, nil
		}
		log.Printf("[AI Orchestrator] Gemini intent extraction failed: %v. Falling back to keyword extraction...", err)
	} else {
		log.Println("[AI Orchestrator] Gemini not available, falling back to keyword extraction...")
	}

	// Step 3: Final fallback — simple keyword extraction
	log.Println("[AI Orchestrator] Using simple keyword extraction as final fallback")
	return simpleKeywordExtract(query), nil
}

// GenerateSummary tries Gemini first (better at natural language), then Grok, then static fallback
func (c *aiClientImpl) GenerateSummary(places []model.Place) (string, error) {
	// Step 1: Try Gemini (preferred for summaries)
	if c.gemini.IsAvailable() {
		summary, err := c.gemini.GenerateSummary(places)
		if err == nil {
			return summary, nil
		}
		log.Printf("[AI Orchestrator] Gemini summary generation failed: %v. Falling back to Grok...", err)
	} else {
		log.Println("[AI Orchestrator] Gemini not available for summary, trying Grok...")
	}

	// Step 2: Try Grok (fallback)
	if c.grok.IsAvailable() {
		summary, err := c.grok.GenerateSummary(places)
		if err == nil {
			return summary, nil
		}
		log.Printf("[AI Orchestrator] Grok summary generation failed: %v. Using static fallback...", err)
	}

	// Step 3: Static fallback
	return generateStaticSummary(places), nil
}

// simpleKeywordExtract performs basic keyword extraction when no AI is available
func simpleKeywordExtract(query string) *model.AIIntent {
	q := strings.ToLower(strings.TrimSpace(query))

	intent := &model.AIIntent{
		Query:    "",
		Location: "",
		Filters:  []string{},
	}

	// Known location keywords (Indonesian cities)
	locations := []string{
		"jakarta", "bekasi", "bandung", "surabaya", "yogyakarta", "semarang",
		"medan", "makassar", "palembang", "tangerang", "depok", "bogor",
		"malang", "solo", "denpasar", "bali",
	}

	for _, loc := range locations {
		if strings.Contains(q, loc) {
			intent.Location = strings.Title(loc)
			break
		}
	}

	// Known category keywords
	categories := map[string]string{
		"makan":      "restaurant",
		"restoran":   "restaurant",
		"restaurant": "restaurant",
		"cafe":       "cafe",
		"kafe":       "cafe",
		"kopi":       "coffee",
		"coffee":     "coffee",
		"taman":      "park",
		"park":       "park",
		"hotel":      "hotel",
		"mall":       "mall",
		"gym":        "gym",
		"bar":        "bar",
		"club":       "nightclub",
	}

	for keyword, category := range categories {
		if strings.Contains(q, keyword) {
			intent.Query = category
			break
		}
	}

	// If no category found, use the whole query
	if intent.Query == "" {
		intent.Query = q
	}

	// Known filter keywords
	filterMap := map[string]string{
		"murah":    "cheap",
		"cheap":    "cheap",
		"mahal":    "expensive",
		"wifi":     "wifi",
		"parkir":   "parking",
		"halal":    "halal",
		"24 jam":   "24 hours",
		"24jam":    "24 hours",
		"outdoor":  "outdoor",
		"indoor":   "indoor",
		"romantis": "romantic",
	}

	for keyword, filter := range filterMap {
		if strings.Contains(q, keyword) {
			intent.Filters = append(intent.Filters, filter)
		}
	}

	log.Printf("[Keyword Extractor] Extracted: query=%s, location=%s, filters=%v", intent.Query, intent.Location, intent.Filters)
	return intent
}

// generateStaticSummary creates a basic summary when AI summary generation is unavailable
func generateStaticSummary(places []model.Place) string {
	if len(places) == 0 {
		return "No places found matching your search."
	}
	if len(places) == 1 {
		return "We found 1 place that matches your search."
	}
	return "Here are some great places based on your search!"
}
