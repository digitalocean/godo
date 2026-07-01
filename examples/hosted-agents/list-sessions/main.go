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
	client := mustClient()
	ctx := context.Background()

	opt := &godo.HostedAgentSessionListOptions{PageSize: 50}
	if v := os.Getenv("SESSION_STATUS"); v != "" {
		opt.Status = godo.HostedAgentSessionStatus(v)
	}

	out, resp, err := client.HostedAgents.ListSessions(ctx, opt)
	if err != nil {
		die(err)
	}

	fmt.Printf("HTTP %d — %d session(s)\n", resp.StatusCode, len(out.Sessions))
	for _, s := range out.Sessions {
		fmt.Printf("  %s  status=%s  agent=%s\n",
			s.SessionID, s.Status, s.AgentKind)
	}
	if out.NextPageToken != "" {
		fmt.Printf("next_page_token: %s\n", out.NextPageToken)
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
