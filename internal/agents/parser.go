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

type Parser struct{}

func NewParser() *Parser {
	_ = godotenv.Load()
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
		// A fix mock tojás és kenyér bevásárlólista
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
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return models.ShoppingList{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return models.ShoppingList{}, fmt.Errorf("Gemini API timeout or connection error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.ShoppingList{}, fmt.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return models.ShoppingList{}, fmt.Errorf("failed to decode Gemini API response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return models.ShoppingList{}, fmt.Errorf("invalid or empty response from Gemini API")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text

	var shoppingList models.ShoppingList
	if err := json.Unmarshal([]byte(responseText), &shoppingList); err != nil {
		return models.ShoppingList{}, fmt.Errorf("failed to parse shopping list JSON from response text: %w", err)
	}

	return shoppingList, nil
}
