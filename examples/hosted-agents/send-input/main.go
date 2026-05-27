package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	sessionID := os.Getenv("HOSTED_AGENT_SESSION_ID")
	if sessionID == "" {
		fmt.Fprintln(os.Stderr, "HOSTED_AGENT_SESSION_ID is required")
		os.Exit(2)
	}

	text := os.Getenv("INPUT_TEXT")
	if text == "" {
		fmt.Fprintln(os.Stderr, "INPUT_TEXT is required")
		os.Exit(2)
	}

	client := mustClient()
	ctx := context.Background()

	out, resp, err := client.HostedAgents.SendInput(ctx, sessionID, &godo.HostedAgentSendInputRequest{
		Text: text,
	})
	if err != nil {
		die(err)
	}

	fmt.Printf("HTTP %d\n", resp.StatusCode)
	fmt.Printf("session_id: %s\n", sessionID)
	fmt.Printf("run_id:     %s\n", out.RunID)
	fmt.Println("\nWatch the run with:")
	fmt.Printf("  go run ./examples/hosted-agents/stream-session\n")
}

func mustClient() *godo.Client {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "DIGITALOCEAN_TOKEN is required")
		os.Exit(2)
	}
	client := godo.NewFromToken(token)
	if baseURL := os.Getenv("DIGITALOCEAN_API_URL"); baseURL != "" {
		u, err := url.Parse(baseURL)
		if err != nil {
			panic(err)
		}
		client.BaseURL = u
	}
	return client
}

func die(err error) {
	var apiErr *godo.ErrorResponse
	if errors.As(err, &apiErr) {
		fmt.Fprintf(os.Stderr, "API error (HTTP %d): %s\n", apiErr.Response.StatusCode, apiErr.Message)
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(1)
}
