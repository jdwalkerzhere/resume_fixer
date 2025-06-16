package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"resume_fixer/prompts"
	"resume_fixer/utils"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/jdwalkerzhere/agent"
	"github.com/jdwalkerzhere/agent/clients"
	"github.com/joho/godotenv"
)

func getInterviewPath() string {
	fmt.Println("Please select the type of interview:")
	i := 1
	desiredPrompt := ""
	for {
		for name := range prompts.PromptRegistry {
			fmt.Printf("\t-%d. %s\n", i, name)
			i++
		}
		fmt.Print("Selection (Type Word, Not Number): ")
		scanForInput := bufio.NewScanner(os.Stdin)
		if scanForInput.Scan() {
			selected := scanForInput.Text()
			_, ok := prompts.PromptRegistry[selected]
			if !ok {
				fmt.Println("Please select a valid interview type.")
				i = 1
			} else {
				desiredPrompt = selected
				break
			}
		}
	}

	selectedPrompt := prompts.PromptRegistry[desiredPrompt].PathToPrompt
	return fmt.Sprintf("./prompts/%s", selectedPrompt)
}

func main() {
	pathToPrompt := getInterviewPath()
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Warning: Error loading .env file: %s\n", err)
	}

	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable is required")
		os.Exit(1)
	}

	// Create a custom HTTP client with enhanced cancellation support
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		DisableKeepAlives:     false, // Enable keep-alives for connection reuse
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Initialize the client with the API key and custom HTTP client
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(httpClient),
	)

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}

		return scanner.Text(), true
	}

	// Create a ReadFileInput struct for the system prompt file
	systemPrompt, err := utils.ReadFileUtil(pathToPrompt)
	if err != nil {
		fmt.Printf("Error finding the promptFilePath: %s", pathToPrompt)
		os.Exit(1)
	}

	// Create context with cancellation support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Create the agent instance
	// Define available tools
	tools := []agent.ToolDefinition{agent.ReadFileDefinition, agent.ListFilesDefinition, agent.EditFileDefinition}
	agentInstance := agent.NewAgent(client, getUserMessage, systemPrompt, tools)

	// Create and start the console client
	console := clients.NewConsoleClient(os.Stdout)
	go console.Run(agentInstance)

	// Handle shutdown signals
	go func() {
		<-signalCh
		fmt.Println("\nShutting down...")
		agentInstance.Close() // Close the agent which will close the output channel
		cancel()

		// Start a cleanup goroutine that will force exit if graceful shutdown takes too long
		go func() {
			// Wait for a short timeout to allow for graceful shutdown
			shutdownTimeout := 1 * time.Second
			time.Sleep(shutdownTimeout)

			// If we're still running after the timeout, force exit
			fmt.Println("Forced shutdown after timeout")
			os.Exit(0)
		}()
	}()

	err = agentInstance.Run(ctx)
	if err != nil {
		fmt.Printf("Error running agent: %s\n", err)
		os.Exit(1)
	}
}
