package agents

import "smart-shopper-agent/internal/models"

const ParserSystemPrompt = "You are a shopping assistant parser that extracts shopping items and quantities from natural language user input."

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(input string) (models.ShoppingList, error) {
	// A fix mock tojás és kenyér bevásárlólista
	return models.ShoppingList{
		Items: []models.ShoppingItem{
			{Name: "tojás", Quantity: 10},
			{Name: "kenyér", Quantity: 1},
		},
	}, nil
}

