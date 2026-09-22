package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	sessionID := os.Getenv("HOSTED_AGENT_SESSION_ID")
	requestID := os.Getenv("HITL_REQUEST_ID")
	if sessionID == "" || requestID == "" {
		fmt.Fprintln(os.Stderr, "HOSTED_AGENT_SESSION_ID and HITL_REQUEST_ID are required")
		os.Exit(2)
	}

	outcome := godo.HostedAgentHITLOutcomeApprove
	if v := os.Getenv("HITL_OUTCOME"); v != "" {
		outcome = godo.HostedAgentHITLOutcome(v)
	}

	// HITL_CONTENT, if set, must be a JSON object — the answer to an MCP form
	// elicitation (e.g. {"site_url": "https://acme.atlassian.net"}). Leave it
	// unset for a plain approval, where Outcome alone is the whole answer.
	var content map[string]any
	if v := os.Getenv("HITL_CONTENT"); v != "" {
		if err := json.Unmarshal([]byte(v), &content); err != nil {
			fmt.Fprintf(os.Stderr, "HITL_CONTENT must be a JSON object: %v\n", err)
			os.Exit(2)
		}
	}

	client := mustClient()
	ctx := context.Background()

	resp, err := client.HostedAgents.ResolveHITL(ctx, sessionID, requestID, &godo.HostedAgentResolveHITLRequest{
		Outcome: outcome,
		Reason:  os.Getenv("HITL_REASON"),
		Source:  godo.HostedAgentResolutionSourceOutOfBand,
		Content: content,
	})
	if err != nil {
		die(err)
	}

	fmt.Printf("HTTP %d — HITL %s resolved with %s\n", resp.StatusCode, requestID, outcome)
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
