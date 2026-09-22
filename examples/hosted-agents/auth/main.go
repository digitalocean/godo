package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

func main() {
	provider := os.Getenv("HOSTED_AGENT_PROVIDER")
	if provider == "" {
		provider = "github"
	}

	client := mustClient()
	ctx := context.Background()

	start, _, err := client.HostedAgents.StartProviderAuth(ctx, provider)
	if err != nil {
		die(err)
	}

	if start.Status == "success" {
		fmt.Printf("%s is already connected for your team\n", provider)
		return
	}

	fmt.Printf("To connect %s, open this URL and authorize access:\n\n  %s\n\n", provider, start.ConnectURL)
	if start.VerificationCode != "" {
		fmt.Printf("Verify the page shows code: %s\n\n", start.VerificationCode)
	}

	fmt.Println("Waiting for authorization to complete...")
	for {
		poll, _, err := client.HostedAgents.PollProviderAuth(ctx, provider, start.PollURL)
		if err != nil {
			die(err)
		}
		if poll.Status == "success" {
			fmt.Printf("%s connected successfully\n", provider)
			return
		}
		time.Sleep(2 * time.Second)
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
