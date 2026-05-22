// Command delete-model-evaluation-preset deletes a saved model evaluation
// preset via the GradientAI API.
//
// Required env vars:
//   - DIGITALOCEAN_TOKEN: a DigitalOcean API token.
//   - DIGITALOCEAN_EVAL_PRESET_UUID: UUID of the preset to delete.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))

	presetUUID := os.Getenv("DIGITALOCEAN_EVAL_PRESET_UUID")
	if presetUUID == "" {
		fmt.Fprintln(os.Stderr, "DIGITALOCEAN_EVAL_PRESET_UUID is required")
		os.Exit(1)
	}

	ctx := context.Background()

	_, resp, err := client.GradientAI.DeleteModelEvaluationPreset(ctx, presetUUID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Deleted model evaluation preset %s (status=%d)\n", presetUUID, resp.Response.StatusCode)
}
