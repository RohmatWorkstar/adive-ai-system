package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"map-backend/internal/model"
)

// GeminiClient handles all interactions with the Gemini API (via OpenAI-compatible endpoint)
type GeminiClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewGeminiClient creates a new Gemini client with a reusable HTTP client
func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsAvailable returns true if the client has a valid API key configured
func (c *GeminiClient) IsAvailable() bool {
	return c.apiKey != ""
}

// makeRequest sends a chat completion request to the Gemini OpenAI-compatible API
func (c *GeminiClient) makeRequest(systemPrompt, userPrompt string, useJSON bool) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"

	reqBody := map[string]interface{}{
		"model": "gemini-2.0-flash",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	if useJSON {
		reqBody["response_format"] = map[string]interface{}{
			"type": "json_object",
		}
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("gemini: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini: API error status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("gemini: failed to decode response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", errors.New("gemini: no choices returned")
	}

	choiceMap, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", errors.New("gemini: invalid choice structure")
	}
	message, ok := choiceMap["message"].(map[string]interface{})
	if !ok {
		return "", errors.New("gemini: invalid message structure")
	}
	content, ok := message["content"].(string)
	if !ok {
		return "", errors.New("gemini: invalid content structure")
	}

	return content, nil
}

// ExtractIntent uses Gemini to extract structured search intent from a natural language query
func (c *GeminiClient) ExtractIntent(query string) (*model.AIIntent, error) {
	if !c.IsAvailable() {
		return nil, errors.New("gemini: API key not configured")
	}

	systemPrompt := `You are an AI assistant that extracts structure from natural language map queries.
Return a strict JSON object with:
- "query": (string) The main search keyword (e.g., "restaurant", "cafe", "park").
- "location": (string) The specific city/area mentioned. If none, leave empty string.
- "filters": (array of strings) Any specific conditions like "cheap", "wifi", "quiet", "halal".
Do NOT include any text outside the JSON object.`

	log.Println("[Gemini] Attempting intent extraction...")
	content, err := c.makeRequest(systemPrompt, query, true)
	if err != nil {
		return nil, err
	}

	var intent model.AIIntent
	if err := json.Unmarshal([]byte(content), &intent); err != nil {
		return nil, fmt.Errorf("gemini: failed to parse AI response as JSON: %w (raw: %s)", err, content)
	}

	log.Printf("[Gemini] Intent extracted: query=%s, location=%s, filters=%v", intent.Query, intent.Location, intent.Filters)
	return &intent, nil
}

// GenerateSummary uses Gemini to generate a human-friendly summary of place results
func (c *GeminiClient) GenerateSummary(places []model.Place) (string, error) {
	if !c.IsAvailable() {
		return "", errors.New("gemini: API key not configured")
	}

	systemPrompt := `You are a helpful travel assistant. Given a list of top places found, write a 1-2 sentence compelling summary of what the user can expect. Be friendly and informative.`

	placesJSON, _ := json.Marshal(places)
	userPrompt := "Here are the top places:\n" + string(placesJSON)

	log.Println("[Gemini] Attempting summary generation...")
	content, err := c.makeRequest(systemPrompt, userPrompt, false)
	if err != nil {
		return "", err
	}

	log.Println("[Gemini] Summary generated successfully.")
	return content, nil
}
