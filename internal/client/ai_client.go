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

type AIClient interface {
	ExtractIntent(query string) (*model.AIIntent, error)
	GenerateSummary(places []model.Place) (string, error)
}

type aiClientImpl struct {
	groqApiKey   string
	geminiApiKey string
	httpClient   *http.Client
}

func NewAIClient(groqApiKey, geminiApiKey string) AIClient {
	return &aiClientImpl{
		groqApiKey:   groqApiKey,
		geminiApiKey: geminiApiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // Timeout to prevent hanging
		},
	}
}

// executeRequest sends an HTTP request and checks the status
func (c *aiClientImpl) executeRequest(url, apiKey string, reqBody map[string]interface{}) (string, error) {
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("api error: status %d, body %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", errors.New("no choices returned by API")
	}

	messageMap, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", errors.New("invalid choice structure")
	}
	message, ok := messageMap["message"].(map[string]interface{})
	if !ok {
		return "", errors.New("invalid message structure")
	}
	content, ok := message["content"].(string)
	if !ok {
		return "", errors.New("invalid content structure")
	}

	return content, nil
}

func (c *aiClientImpl) makeRequestWithFailover(systemPrompt, userPrompt string, useJSON bool) (string, error) {
	// 1. Try Groq as Primary
	groqUrl := "https://api.groq.com/openai/v1/chat/completions"
	groqBody := map[string]interface{}{
		"model": "llama3-70b-8192", 
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	if useJSON {
		groqBody["response_format"] = map[string]interface{}{
			"type": "json_object",
		}
	}

	log.Println("Attempting AI request with Groq...")
	content, err := c.executeRequest(groqUrl, c.groqApiKey, groqBody)
	if err == nil {
		log.Println("Groq request successful.")
		return content, nil
	}

	log.Printf("Groq API request failed: %v. Initiating fallback to Gemini...\n", err)

	// 2. Try Gemini as Fallback (using OpenAI compatibility layer)
	geminiUrl := "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"
	geminiBody := map[string]interface{}{
		"model": "gemini-1.5-flash",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	if useJSON {
		geminiBody["response_format"] = map[string]interface{}{
			"type": "json_object",
		}
	}

	contentFallback, errFallback := c.executeRequest(geminiUrl, c.geminiApiKey, geminiBody)
	if errFallback == nil {
		log.Println("Gemini fallback request successful.")
		return contentFallback, nil
	}

	return "", fmt.Errorf("all AI APIs failed. Primary Error: %v, Fallback Error: %v", err, errFallback)
}

func (c *aiClientImpl) ExtractIntent(query string) (*model.AIIntent, error) {
	systemPrompt := `You are an AI assistant that extracts structure from natural language map queries.
Return a strict JSON object with:
- "location": (string) The specific city/area mentioned. If none, leave empty string.
- "category": (string) The main type of place (e.g., "restaurant", "cafe", "park").
- "filters": (array of strings) Any specific conditions like "cheap", "wifi", "quiet".`

	content, err := c.makeRequestWithFailover(systemPrompt, query, true)
	if err != nil {
		return nil, err
	}

	var intent model.AIIntent
	if err := json.Unmarshal([]byte(content), &intent); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w. raw content: %s", err, content)
	}

	return &intent, nil
}

func (c *aiClientImpl) GenerateSummary(places []model.Place) (string, error) {
	systemPrompt := `You are a helpful travel assistant. Given a list of top places found, write a 1-2 sentence compelling summary of what the user can expect.`

	placesJSON, _ := json.Marshal(places)
	userPrompt := "Here are the top places:\n" + string(placesJSON)

	content, err := c.makeRequestWithFailover(systemPrompt, userPrompt, false)
	if err != nil {
		return "", err
	}

	return content, nil
}
