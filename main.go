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

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/joho/godotenv"
	"resume_fixer/agent"
)

func main() {
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

	interviewPrompt := "./agent_config/system_prompts/candidate_discovery.md"
	systemPrompt, err := agent.ReadFile(interviewPrompt)
	if err != nil {
		fmt.Printf("No system prompt for the agent found: %s\n", err)
		os.Exit(1)
	}

	// Create context with cancellation support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Create the agent instance
	agentInstance := agent.NewAgent(client, getUserMessage, systemPrompt)

	// Create and start the console client
	console := agent.NewConsoleClient(os.Stdout)
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
			shutdownTimeout := 2 * time.Second
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
