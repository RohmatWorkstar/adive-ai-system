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

// GrokClient handles all interactions with the Groq API (Grok provider)
type GrokClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewGrokClient creates a new Grok client with a reusable HTTP client
func NewGrokClient(apiKey string) *GrokClient {
	return &GrokClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsAvailable returns true if the client has a valid API key configured
func (c *GrokClient) IsAvailable() bool {
	return c.apiKey != ""
}

// makeRequest sends a chat completion request to the Groq API
func (c *GrokClient) makeRequest(systemPrompt, userPrompt string, useJSON bool) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": "llama3-70b-8192",
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
		return "", fmt.Errorf("grok: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("grok: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("grok: API error status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("grok: failed to decode response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", errors.New("grok: no choices returned")
	}

	choiceMap, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", errors.New("grok: invalid choice structure")
	}
	message, ok := choiceMap["message"].(map[string]interface{})
	if !ok {
		return "", errors.New("grok: invalid message structure")
	}
	content, ok := message["content"].(string)
	if !ok {
		return "", errors.New("grok: invalid content structure")
	}

	return content, nil
}

// ExtractIntent uses Grok to extract structured search intent from a natural language query
func (c *GrokClient) ExtractIntent(query string) (*model.AIIntent, error) {
	if !c.IsAvailable() {
		return nil, errors.New("grok: API key not configured")
	}

	systemPrompt := `You are an AI assistant that extracts structure from natural language map queries.
Return a strict JSON object with:
- "query": (string) The main search keyword (e.g., "restaurant", "cafe", "park").
- "location": (string) The specific city/area mentioned. If none, leave empty string.
- "filters": (array of strings) Any specific conditions like "cheap", "wifi", "quiet", "halal".
Do NOT include any text outside the JSON object.`

	log.Println("[Grok] Attempting intent extraction...")
	content, err := c.makeRequest(systemPrompt, query, true)
	if err != nil {
		return nil, err
	}

	var intent model.AIIntent
	if err := json.Unmarshal([]byte(content), &intent); err != nil {
		return nil, fmt.Errorf("grok: failed to parse AI response as JSON: %w (raw: %s)", err, content)
	}

	log.Printf("[Grok] Intent extracted: query=%s, location=%s, filters=%v", intent.Query, intent.Location, intent.Filters)
	return &intent, nil
}

// GenerateSummary uses Grok to generate a human-friendly summary of place results
func (c *GrokClient) GenerateSummary(places []model.Place) (string, error) {
	if !c.IsAvailable() {
		return "", errors.New("grok: API key not configured")
	}

	systemPrompt := `You are a helpful travel assistant. Given a list of top places found, write a 1-2 sentence compelling summary of what the user can expect. Be friendly and informative.`

	placesJSON, _ := json.Marshal(places)
	userPrompt := "Here are the top places:\n" + string(placesJSON)

	log.Println("[Grok] Attempting summary generation...")
	content, err := c.makeRequest(systemPrompt, userPrompt, false)
	if err != nil {
		return "", err
	}

	log.Println("[Grok] Summary generated successfully.")
	return content, nil
}
