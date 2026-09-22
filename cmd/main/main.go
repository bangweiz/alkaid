package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"charm.land/glamour/v2"
	"github.com/bangweiz/alkaid/internal/provider/gemini"
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
	req := gemini.NewInteractionRequest(gemini.ModelGemini35FlashLite, "How does AI work?")

	interaction, err := client.CreateInteraction(context.Background(), &req)
	if err != nil {
		panic(err)
	}

	for _, step := range interaction.Steps {
		for _, content := range step.Content {
			output, err := glamour.Render(content.Text, "light")
			if err != nil {
				panic(err)
			}
			fmt.Print(output)
		}
	}
}
