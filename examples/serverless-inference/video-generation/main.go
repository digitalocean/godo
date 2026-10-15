// Command video-generation starts an async video job, polls to completion, and
// writes output.mp4 via the content endpoint (or a presigned output URL).
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

func main() {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		token = os.Getenv("MODEL_ACCESS_KEY")
	}
	client := godo.NewFromToken(token)

	timeout := durationFromEnv("DIGITALOCEAN_INFERENCE_TIMEOUT", 10*time.Minute)
	interval := durationFromEnv("DIGITALOCEAN_INFERENCE_POLL_INTERVAL", 5*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	model := os.Getenv("DIGITALOCEAN_INFERENCE_VIDEO_MODEL")
	if model == "" {
		model = "wan2-2-t2v-a14b"
	}

	job, _, err := client.Video.Generations.Create(ctx, &godo.VideoGenerateParams{
		Model:  model,
		Prompt: "Two anthropomorphic cats in boxing gear fight intensely on a spotlighted stage",
		Size:   "1280x720",
		FPS:    godo.PtrTo(16),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Started video job id=%s status=%s\n", job.ID, job.Status)

	final, err := poll(ctx, client, job.ID, interval)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Terminal status=%s id=%s\n", final.Status, final.ID)
	if final.Status != "completed" {
		panic(fmt.Sprintf("video generation ended with status %q", final.Status))
	}

	dest := "./output.mp4"
	if final.Output != nil && final.Output.URL != "" {
		if err := downloadURL(ctx, final.Output.URL, dest); err != nil {
			panic(err)
		}
	} else {
		raw, _, err := client.Video.Generations.Content(ctx, final.ID)
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(dest, raw, 0644); err != nil {
			panic(err)
		}
	}
	fmt.Println("Wrote", dest)
}

func poll(ctx context.Context, client *godo.Client, id string, interval time.Duration) (*godo.VideoGeneration, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		job, _, err := client.Video.Generations.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		fmt.Printf("  [%s] status=%s id=%s\n", time.Now().Format("15:04:05"), job.Status, job.ID)
		switch job.Status {
		case "completed", "failed":
			return job, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("polling %s: %w (last status %q)", id, ctx.Err(), job.Status)
		case <-ticker.C:
		}
	}
}

func downloadURL(ctx context.Context, rawURL, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("download %s: %s: %s", rawURL, resp.Status, body)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func durationFromEnv(name string, fallback time.Duration) time.Duration {
	if v := os.Getenv(name); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
