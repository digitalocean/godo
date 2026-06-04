// Command get-model-evaluation-run-results-download-url returns a presigned
// download URL for a model evaluation run's results (gzip-compressed JSON) via
// the GradientAI API.
//
// Required env vars:
//   - DIGITALOCEAN_TOKEN: a DigitalOcean API token.
//   - DIGITALOCEAN_EVAL_RUN_UUID: UUID of the evaluation run.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))

	runUUID := os.Getenv("DIGITALOCEAN_EVAL_RUN_UUID")
	if runUUID == "" {
		fmt.Fprintln(os.Stderr, "DIGITALOCEAN_EVAL_RUN_UUID is required")
		os.Exit(1)
	}

	ctx := context.Background()

	out, _, err := client.GradientAI.GetModelEvaluationRunResultsDownloadURL(ctx, runUUID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Download URL: %s\n", out.DownloadURL)
	if out.ExpiresAt != nil {
		fmt.Printf("Expires at:   %s\n", out.ExpiresAt)
	}
}
