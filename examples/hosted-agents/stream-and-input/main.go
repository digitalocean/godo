package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

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

	text := envOr("INPUT_TEXT", "What files are in this repository?")
	out, _, err := client.HostedAgents.SendInput(ctx, sessionID, &godo.HostedAgentSendInputRequest{
		Text: text,
	})
	if err != nil {
		die(err)
	}
	fmt.Printf("SendInput → run_id=%s\n\n", out.RunID)

	streamCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	stream, _, err := client.HostedAgents.StreamSession(streamCtx, sessionID, nil)
	if err != nil {
		die(err)
	}
	defer stream.Close()

	fmt.Println("StreamSession events:")
	for stream.Next() {
		ev := stream.Current()
		payload := string(ev.Payload)
		if len(payload) > 120 {
			payload = payload[:120] + "..."
		}
		fmt.Printf("  [%s] %s  payload=%s\n", ev.EventID, ev.Kind, payload)

		if ev.Kind == godo.HostedAgentEventKindTokenChunk && len(ev.Payload) > 0 {
			var p struct {
				Text string `json:"text"`
			}
			if json.Unmarshal(ev.Payload, &p) == nil && p.Text != "" {
				fmt.Print(p.Text)
			}
		}
	}
	if err := stream.Err(); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		die(err)
	}
	fmt.Println()
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

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
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
