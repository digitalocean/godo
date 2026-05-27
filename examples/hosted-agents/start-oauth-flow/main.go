package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/digitalocean/godo"
)

func main() {
	sessionID := os.Getenv("HOSTED_AGENT_SESSION_ID")
	if sessionID == "" {
		fmt.Fprintln(os.Stderr, "HOSTED_AGENT_SESSION_ID is required")
		os.Exit(2)
	}

	provider := envOr("OAUTH_PROVIDER", "github")

	var req *godo.HostedAgentStartOAuthFlowRequest
	if scopes := os.Getenv("OAUTH_SCOPES"); scopes != "" {
		req = &godo.HostedAgentStartOAuthFlowRequest{
			RequestedScopes: strings.Split(scopes, ","),
		}
	}

	client := mustClient()
	ctx := context.Background()

	out, resp, err := client.HostedAgents.StartOAuthFlow(ctx, sessionID, provider, req)
	if err != nil {
		die(err)
	}

	fmt.Printf("HTTP %d\n", resp.StatusCode)
	fmt.Printf("flow_kind:     %s\n", out.FlowKind)
	fmt.Printf("authorize_url: %s\n", out.AuthorizeURL)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
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
