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

	limit := 0
	if v := os.Getenv("LIMIT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			fmt.Fprintln(os.Stderr, "LIMIT must be a positive integer")
			os.Exit(2)
		}
		limit = n
	}

	var opt *godo.HostedAgentSessionStreamOptions
	replayFrom, before := os.Getenv("REPLAY_FROM"), os.Getenv("BEFORE")
	replayOnly := os.Getenv("REPLAY_ONLY") == "true"
	if replayFrom != "" || before != "" || limit > 0 || replayOnly {
		opt = &godo.HostedAgentSessionStreamOptions{
			ReplayFrom: replayFrom,
			ReplayOnly: replayOnly,
			Before:     before,
			Limit:      limit,
		}
	}

	stream, resp, err := client.HostedAgents.StreamSession(streamCtx, sessionID, opt)
	if err != nil {
		die(err)
	}
	defer stream.Close()

	fmt.Printf("HTTP %d — streaming session %s (%s timeout)\n\n", resp.StatusCode, sessionID, timeout)

	var count int
	var oldest string
	for stream.Next() {
		ev := stream.Current()
		// Events arrive oldest-first, so the first one seen is the cursor to
		// page further back from.
		if count == 0 {
			oldest = ev.EventID
		}
		count++
		payload := string(ev.Payload)
		if len(payload) > 160 {
			payload = payload[:160] + "..."
		}
		fmt.Printf("[%s] %s run=%s\n  payload=%s\n\n", ev.EventID, ev.Kind, ev.RunID, payload)
	}
	if err := stream.Err(); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		die(err)
	}

	fmt.Printf("%d events, oldest=%q, has_more=%t\n", count, oldest, stream.HasMore())
	if stream.HasMore() {
		fmt.Printf("older history remains: rerun with BEFORE=%s\n", oldest)
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
