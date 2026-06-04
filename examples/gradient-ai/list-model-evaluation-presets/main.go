// Command list-model-evaluation-presets lists all saved model evaluation
// presets via the GradientAI API.
//
// Required env vars:
//   - DIGITALOCEAN_TOKEN: a DigitalOcean API token.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))

	ctx := context.Background()

	out, _, err := client.GradientAI.ListModelEvaluationPresets(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d model evaluation presets\n", len(out.Presets))
	for _, preset := range out.Presets {
		fmt.Printf("- %s (%s) dataset=%s judge=%s metrics=%d\n",
			preset.EvalPresetUuid, preset.Name, preset.DatasetName, preset.JudgeModelName, len(preset.Metrics))
	}
}
