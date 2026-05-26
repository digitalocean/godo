// Command list-model-evaluation-runs lists model evaluation runs, optionally
// filtered by preset UUID and status, via the GradientAI API.
//
// Required env vars:
//   - DIGITALOCEAN_TOKEN: a DigitalOcean API token.
//
// Optional env vars:
//   - DIGITALOCEAN_EVAL_PRESET_UUID: filter results to runs from this preset.
//   - DIGITALOCEAN_EVAL_RUN_STATUS: filter by run status (e.g.
//     MODEL_EVALUATION_RUN_SUCCESSFUL).
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))

	opt := &godo.ModelEvaluationRunListOptions{
		EvalPresetUUID: os.Getenv("DIGITALOCEAN_EVAL_PRESET_UUID"),
		Status:         godo.ModelEvaluationRunStatus(os.Getenv("DIGITALOCEAN_EVAL_RUN_STATUS")),
	}

	ctx := context.Background()

	out, _, err := client.GradientAI.ListModelEvaluationRuns(ctx, opt)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d model evaluation runs\n", len(out.Runs))
	for _, run := range out.Runs {
		fmt.Printf("- %s (%s) status=%s candidate=%s judge=%s\n",
			run.EvalRunUuid, run.Name, run.Status, run.CandidateModelName, run.JudgeModelName)
	}
}
