package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	timeout := 60 * time.Second
	if v := os.Getenv("STREAM_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeout = time.Duration(n) * time.Second
		}
	}
	streamCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var opt *godo.HostedAgentSessionStreamOptions
	if from := os.Getenv("REPLAY_FROM"); from != "" || os.Getenv("REPLAY_ONLY") == "true" {
		opt = &godo.HostedAgentSessionStreamOptions{
			ReplayFrom: os.Getenv("REPLAY_FROM"),
			ReplayOnly: os.Getenv("REPLAY_ONLY") == "true",
		}
	}

	stream, resp, err := client.HostedAgents.StreamSession(streamCtx, sessionID, opt)
	if err != nil {
		die(err)
	}
	defer stream.Close()

	fmt.Printf("HTTP %d — streaming session %s (%s timeout)\n\n", resp.StatusCode, sessionID, timeout)

	for stream.Next() {
		ev := stream.Current()
		payload := string(ev.Payload)
		if len(payload) > 160 {
			payload = payload[:160] + "..."
		}
		fmt.Printf("[%s] %s run=%s\n  payload=%s\n\n", ev.EventID, ev.Kind, ev.RunID, payload)
	}
	if err := stream.Err(); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		die(err)
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
