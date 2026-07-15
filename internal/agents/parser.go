package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"smart-shopper-agent/internal/models"
	"time"

	"github.com/joho/godotenv"
)

const ParserSystemPrompt = "You are a shopping assistant parser that extracts shopping items and quantities from natural language user input."

type Parser struct {
	Client *http.Client
	APIURL string
}

func NewParser() *Parser {
	_ = godotenv.Load()
	return &Parser{
		Client: &http.Client{Timeout: 10 * time.Second},
		APIURL: "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent",
	}
}

type Part struct {
	Text string `json:"text"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type SystemInstruction struct {
	Parts []Part `json:"parts"`
}

type GenerationConfig struct {
	ResponseMimeType string `json:"responseMimeType"`
}

type GeminiRequest struct {
	Contents          []Content         `json:"contents"`
	SystemInstruction SystemInstruction `json:"systemInstruction"`
	GenerationConfig  GenerationConfig  `json:"generationConfig"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func buildRequestBody(input string) ([]byte, error) {
	reqBody := GeminiRequest{
		Contents:          []Content{{Parts: []Part{{Text: input}}}},
		SystemInstruction: SystemInstruction{Parts: []Part{{Text: ParserSystemPrompt}}},
		GenerationConfig:  GenerationConfig{ResponseMimeType: "application/json"},
	}
	return json.Marshal(reqBody)
}

func (p *Parser) doAttempt(client *http.Client, apiURL, apiKey string, jsonData []byte, attempt int) (models.ShoppingList, error) {
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return models.ShoppingList{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return models.ShoppingList{}, fmt.Errorf("Gemini API network error on attempt %d: %w", attempt+1, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.ShoppingList{}, fmt.Errorf("API request failed with status code %d on attempt %d", resp.StatusCode, attempt+1)
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return models.ShoppingList{}, fmt.Errorf("failed to decode response on attempt %d: %w", attempt+1, err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return models.ShoppingList{}, fmt.Errorf("invalid or empty response from Gemini API on attempt %d", attempt+1)
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	var shoppingList models.ShoppingList
	if err := json.Unmarshal([]byte(responseText), &shoppingList); err != nil {
		return models.ShoppingList{}, fmt.Errorf("failed to parse shopping list JSON from response text on attempt %d: %w", attempt+1, err)
	}
	return shoppingList, nil
}

func (p *Parser) Parse(input string) (models.ShoppingList, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" || apiKey == "your_api_key_here" {
		return models.ShoppingList{}, fmt.Errorf("GEMINI_API_KEY is not set or invalid")
	}

	jsonData, err := buildRequestBody(input)
	if err != nil {
		return models.ShoppingList{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	client := p.Client
	if client == nil {
		client = &http.Client{
			Timeout:   10 * time.Second,
			Transport: http.DefaultTransport,
		}
	}

	apiURL := p.APIURL
	if apiURL == "" {
		apiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"
	}

	var lastErr error
	maxRetries := 2
	baseDelay := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(baseDelay * time.Duration(1<<(attempt-1)))
		}
		shoppingList, err := p.doAttempt(client, apiURL, apiKey, jsonData, attempt)
		if err != nil {
			lastErr = err
			continue
		}
		return shoppingList, nil
	}
	return models.ShoppingList{}, fmt.Errorf("failed to parse after %d retries: %w", maxRetries, lastErr)
}
