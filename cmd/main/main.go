package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	eventChan, err := client.CreateInteraction(context.Background(), &req)
	if err != nil {
		panic(err)
	}

	for result := range eventChan {
		if result.Error != nil {
			log.Printf("Stream error: %v\n", result.Error)
			break
		}

		switch ev := result.Event.(type) {
		case *gemini.StepDeltaEvent:
			if ev.Delta.Type == "text" {
				fmt.Print(ev.Delta.Text)
			}
		case *gemini.InteractionCompletedEvent:
			fmt.Printf("\n\nFinished! Tokens: %d\n", ev.Interaction.Usage.TotalTokens)
		case *gemini.StreamErrorEvent:
			log.Printf("\nAPI Error: %s\n", ev.Error.Message)
		}
	}
}
