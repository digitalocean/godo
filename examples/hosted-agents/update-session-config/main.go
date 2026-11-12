package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

// Usage: HOSTED_AGENT_SESSION_ID=... HOSTED_AGENT_MANIFEST_YAML=$'tools:\n  - ...' go run .
func main() {
	sessionID := os.Getenv("HOSTED_AGENT_SESSION_ID")
	manifest := os.Getenv("HOSTED_AGENT_MANIFEST_YAML")
	if sessionID == "" || manifest == "" {
		fmt.Fprintln(os.Stderr, "HOSTED_AGENT_SESSION_ID and HOSTED_AGENT_MANIFEST_YAML are required")
		os.Exit(2)
	}

	client := mustClient()
	ctx := context.Background()

	res, resp, err := client.HostedAgents.UpdateSessionConfig(ctx, sessionID, &godo.HostedAgentSessionConfigUpdateRequest{ManifestYAML: manifest})
	if err != nil {
		die(err)
	}
	fmt.Printf("HTTP %d outcome=%s\n", resp.StatusCode, res.Outcome)
	for _, w := range res.Warnings {
		fmt.Println("warning:", w)
	}
	if res.Outcome != godo.HostedAgentSessionConfigUpdateOutcomeAccepted {
		return
	}

	// godo does not wait: poll until the session leaves UPDATING.
	for {
		s, _, err := client.HostedAgents.GetSession(ctx, sessionID)
		if err != nil {
			die(err)
		}
		if s.Status != godo.HostedAgentSessionStatusUpdating {
			if s.Update != nil {
				fmt.Printf("update %s: %s %s\n", s.Update.UpdateID, s.Update.State, s.Update.Detail)
			}
			return
		}
		time.Sleep(3 * time.Second)
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
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
