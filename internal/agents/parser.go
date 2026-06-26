package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"smart-shopper-agent/internal/models"
	"time"
)

const ParserSystemPrompt = "You are a shopping assistant parser that extracts shopping items and quantities from natural language user input."

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
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

func (p *Parser) Parse(input string) (models.ShoppingList, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" || apiKey == "your_api_key_here" {
		// Mock response if API key is not available
		return models.ShoppingList{
			Items: []models.ShoppingItem{
				{Name: "tojás", Quantity: 10},
				{Name: "kenyér", Quantity: 1},
			},
		}, nil
	}

	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: input},
				},
			},
		},
		SystemInstruction: SystemInstruction{
			Parts: []Part{
				{Text: ParserSystemPrompt},
			},
		},
		GenerationConfig: GenerationConfig{
			ResponseMimeType: "application/json",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return models.ShoppingList{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", apiKey)

	var lastErr error
	maxRetries := 2
	baseDelay := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(baseDelay * time.Duration(1<<(attempt-1))) // Exponential backoff
		}

		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return models.ShoppingList{}, fmt.Errorf("failed to create HTTP request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("Gemini API network error on attempt %d: %w", attempt+1, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("API request failed with status code %d on attempt %d", resp.StatusCode, attempt+1)
			continue
		}

		var geminiResp GeminiResponse
		err = json.NewDecoder(resp.Body).Decode(&geminiResp)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to decode response on attempt %d: %w", attempt+1, err)
			continue
		}

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			lastErr = fmt.Errorf("invalid or empty response from Gemini API on attempt %d", attempt+1)
			continue
		}

		responseText := geminiResp.Candidates[0].Content.Parts[0].Text

		var shoppingList models.ShoppingList
		if err := json.Unmarshal([]byte(responseText), &shoppingList); err != nil {
			lastErr = fmt.Errorf("failed to parse shopping list JSON from response text on attempt %d: %w", attempt+1, err)
			continue
		}

		// Success
		return shoppingList, nil
	}

	return models.ShoppingList{}, fmt.Errorf("failed to parse after %d retries: %w", maxRetries, lastErr)
}
