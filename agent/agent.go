package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool), systemPrompt string) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		SystemPrompt:   systemPrompt,
	}
}

type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	SystemPrompt   string `json:"system,omitzero"`
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}
	fmt.Printf("\033[31mWhat is the path to your resume?\033[0m: ")
	resumePath, _ := a.getUserMessage()
	resumeContent, err := ReadFile(resumePath)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		os.Exit(1)
	}
	userResume := anthropic.NewUserMessage(anthropic.NewTextBlock(resumeContent))
	conversation = append(conversation, userResume)

	readUserInput := true
	for {
		// Since we're starting with the candidates resume contents
		// We're letting Claude speak first, I need to make this text streaming so it feels more responsive
		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.ToParam())

		for _, content := range message.Content {
			// This is a switch for when I choose to add tools
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
			}
		}

		// Now it's the user's turn
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}

			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
			conversation = append(conversation, userMessage)
		}
	}

	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	// Hacky solution, apparently I can't just send a string over the wire, don't ask
	sysPrompt := []anthropic.TextBlockParam{*anthropic.NewTextBlock(a.SystemPrompt).OfRequestTextBlock}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: int64(1024),
		Messages:  conversation,
		System:    sysPrompt,
	})
	return message, err
}

func ReadFile(path string) (string, error) {
	// Yeah, I should check for a filepath's existance, but I'm not going to
	resume, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(resume), nil
}
