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

	agentKind := godo.HostedAgentKindClaudeCode
	if v := os.Getenv("HOSTED_AGENT_KIND"); v != "" {
		agentKind = godo.HostedAgentKind(v)
	}

	ctx := context.Background()

	if client.HostedAgents == nil {
		fmt.Fprintln(os.Stderr, "HostedAgents service is not initialized on this godo client")
		os.Exit(1)
	}

	session, resp, err := client.HostedAgents.CreateSession(ctx, &godo.HostedAgentSessionCreateRequest{
		AgentKind: agentKind,
		RepoHint:  os.Getenv("REPO_HINT"),
	})
	if err != nil {
		var apiErr *godo.ErrorResponse
		if errors.As(err, &apiErr) {
			fmt.Fprintf(os.Stderr, "API error (HTTP %d): %s\n", apiErr.Response.StatusCode, apiErr.Message)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("HTTP %d\n", resp.StatusCode)
	if session == nil {
		fmt.Fprintln(os.Stderr, "API returned no session")
		os.Exit(1)
	}
	fmt.Printf("session_id:  %s\n", session.SessionID)
	fmt.Printf("status:      %s\n", session.Status)
	fmt.Printf("agent_kind:  %s\n", session.AgentKind)
	if session.RepoHint != "" {
		fmt.Printf("repo_hint:   %s\n", session.RepoHint)
	}
}
