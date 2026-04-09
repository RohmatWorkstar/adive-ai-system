package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"map-backend/internal/model"
)

type OpenAIClient interface {
	ExtractIntent(query string) (*model.AIIntent, error)
	GenerateSummary(places []model.Place) (string, error)
}

type openAIClientImpl struct {
	apiKey     string
	httpClient *http.Client
}

func NewOpenAIClient(apiKey string) OpenAIClient {
	return &openAIClientImpl{
		apiKey: apiKey,
		httpClient: &http.Client{},
	}
}

func (c *openAIClientImpl) makeRequest(systemPrompt, userPrompt string, useJSON bool) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": "gpt-4o-mini", // Cost efficient model
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
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	choices := result["choices"].([]interface{})
	if len(choices) == 0 {
		return "", errors.New("no choice returned by openai")
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	return message["content"].(string), nil
}

func (c *openAIClientImpl) ExtractIntent(query string) (*model.AIIntent, error) {
	systemPrompt := `You are an AI assistant that extracts structure from natural language map queries.
Return a strict JSON object with:
- "location": (string) The specific city/area mentioned. If none, leave empty string.
- "category": (string) The main type of place (e.g., "restaurant", "cafe", "park").
- "filters": (array of strings) Any specific conditions like "cheap", "wifi", "quiet".`

	content, err := c.makeRequest(systemPrompt, query, true)
	if err != nil {
		return nil, err
	}

	var intent model.AIIntent
	if err := json.Unmarshal([]byte(content), &intent); err != nil {
		return nil, err
	}

	return &intent, nil
}

func (c *openAIClientImpl) GenerateSummary(places []model.Place) (string, error) {
	systemPrompt := `You are a helpful travel assistant. Given a list of top places found, write a 1-2 sentence compelling summary of what the user can expect.`

	placesJSON, _ := json.Marshal(places)
	userPrompt := "Here are the top places:\n" + string(placesJSON)

	content, err := c.makeRequest(systemPrompt, userPrompt, false)
	if err != nil {
		return "", err
	}

	return content, nil
}
