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

	client := mustClient()
	ctx := context.Background()

	session, resp, err := client.HostedAgents.GetSession(ctx, sessionID)
	if err != nil {
		die(err)
	}

	fmt.Printf("HTTP %d\n", resp.StatusCode)
	fmt.Printf("session_id:   %s\n", session.SessionID)
	fmt.Printf("status:       %s\n", session.Status)
	fmt.Printf("agent_kind:   %s\n", session.AgentKind)
	fmt.Printf("team_id:      %d\n", session.TeamID)
	fmt.Printf("sandbox_id:   %s\n", session.SandboxID)
	fmt.Printf("repo_hint:    %s\n", session.RepoHint)
	fmt.Printf("created_at:   %s\n", session.CreatedAt)
	fmt.Printf("last_event_at: %s\n", session.LastEventAt)
	if len(session.ProviderAuth) > 0 {
		fmt.Println("provider_auth:")
		for k, v := range session.ProviderAuth {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}
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
