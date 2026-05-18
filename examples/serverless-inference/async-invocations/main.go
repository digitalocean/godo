package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))

	timeout := durationFromEnv("DIGITALOCEAN_INFERENCE_TIMEOUT", 2*time.Minute)
	interval := durationFromEnv("DIGITALOCEAN_INFERENCE_POLL_INTERVAL", 2*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	model := os.Getenv("DIGITALOCEAN_INFERENCE_ASYNC_MODEL")
	if model == "" {
		model = "fal-ai/flux/schnell"
	}

	invocation, _, err := client.AsyncInvocations.New(ctx, &godo.AsyncInvocationNewParams{
		ModelID: model,
		Input: map[string]interface{}{
			"prompt": "A futuristic city at sunset",
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Started async invocation:")
	fmt.Printf("  request_id: %s\n", invocation.RequestID)
	fmt.Printf("  model_id:   %s\n", invocation.ModelID)
	fmt.Printf("  status:     %s\n", invocation.Status)
	fmt.Printf("  created_at: %s\n", invocation.CreatedAt)
	fmt.Println()

	final, err := poll(ctx, client, invocation.RequestID, interval)
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println("Terminal state:")
	fmt.Printf("  request_id:   %s\n", final.RequestID)
	fmt.Printf("  status:       %s\n", final.Status)
	if final.StartedAt != nil {
		fmt.Printf("  started_at:   %s\n", *final.StartedAt)
	}
	if final.CompletedAt != nil {
		fmt.Printf("  completed_at: %s\n", *final.CompletedAt)
	}
	if final.Error != nil {
		fmt.Printf("  error:        %s\n", *final.Error)
	}
	if len(final.Output) > 0 {
		out, _ := json.MarshalIndent(final.Output, "  ", "  ")
		fmt.Printf("  output:       %s\n", out)
	}
}

func poll(ctx context.Context, client *godo.Client, requestID string, interval time.Duration) (*godo.AsyncInvocation, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		inv, _, err := client.AsyncInvocations.Get(ctx, requestID)
		if err != nil {
			return nil, err
		}
		fmt.Printf("  [%s] status=%s\n", time.Now().Format("15:04:05"), inv.Status)
		switch inv.Status {
		case "COMPLETED", "FAILED":
			return inv, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("polling %s: %w (last status %q)", requestID, ctx.Err(), inv.Status)
		case <-ticker.C:
		}
	}
}

func durationFromEnv(name string, fallback time.Duration) time.Duration {
	if v := os.Getenv(name); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
