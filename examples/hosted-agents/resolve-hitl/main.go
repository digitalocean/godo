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
	requestID := os.Getenv("HITL_REQUEST_ID")
	if sessionID == "" || requestID == "" {
		fmt.Fprintln(os.Stderr, "HOSTED_AGENT_SESSION_ID and HITL_REQUEST_ID are required")
		os.Exit(2)
	}

	outcome := godo.HostedAgentHITLOutcomeApprove
	if v := os.Getenv("HITL_OUTCOME"); v != "" {
		outcome = godo.HostedAgentHITLOutcome(v)
	}

	client := mustClient()
	ctx := context.Background()

	resp, err := client.HostedAgents.ResolveHITL(ctx, sessionID, requestID, &godo.HostedAgentResolveHITLRequest{
		Outcome: outcome,
		Reason:  os.Getenv("HITL_REASON"),
		Source:  godo.HostedAgentResolutionSourceOutOfBand,
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
