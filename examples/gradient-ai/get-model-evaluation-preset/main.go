// Command get-model-evaluation-preset retrieves a saved model evaluation
// preset via the GradientAI API.
//
// Required env vars:
//   - DIGITALOCEAN_TOKEN: a DigitalOcean API token.
//   - DIGITALOCEAN_EVAL_PRESET_UUID: UUID of the preset to retrieve.
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

	out, _, err := client.GradientAI.GetModelEvaluationPreset(ctx, presetUUID)
	if err != nil {
		panic(err)
	}

	if out.Preset == nil {
		fmt.Printf("No preset returned for %s\n", presetUUID)
		return
	}

	preset := out.Preset
	fmt.Printf("Model evaluation preset %s\n", preset.EvalPresetUuid)
	fmt.Printf("  name:             %s\n", preset.Name)
	fmt.Printf("  dataset:          %s (%s)\n", preset.DatasetName, preset.DatasetUuid)
	fmt.Printf("  judge model:      %s (%s)\n", preset.JudgeModelName, preset.JudgeModelUuid)
	if preset.CreatedAt != nil {
		fmt.Printf("  created at:       %s\n", preset.CreatedAt)
	}
	for _, m := range preset.Metrics {
		fmt.Printf("    metric %s [%s]\n", m.MetricName, m.MetricUUID)
	}
}
