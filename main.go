package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
	"resume_fixer/agent"
)

func main() {
	godotenv.Load()
	client := anthropic.NewClient()

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}

		return scanner.Text(), true
	}

	systemPrompt, err := agent.ReadFile("systemPrompt.md")
	if err != nil {
		fmt.Printf("No system prompt for the agent found: %s", err)
		os.Exit(1)
	}

	agentInstance := agent.NewAgent(&client, getUserMessage, systemPrompt)
	err = agentInstance.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}
