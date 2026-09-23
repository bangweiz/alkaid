package main

import (
	"context"
	"log"
	"os"

	"github.com/bangweiz/alkaid/internal/agents"
	"github.com/bangweiz/alkaid/internal/llm/gemini"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, relying on system environment variables")
	}

	url := os.Getenv("GEMINI_URL")
	apiKey := os.Getenv("GEMINI_KEY")

	client := gemini.NewClient(url, apiKey)
	agent := agents.NewCodingAgent(client)

	if err := agent.Prompt(context.Background(), "How does Go handle garbage collection?"); err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}
}
