// Command cancel-model-evaluation-run cancels an in-progress model evaluation
// run via the GradientAI API. The run must be in a non-terminal status
// (queued, running_dataset, or evaluating_results).
//
// Required env vars:
//   - DIGITALOCEAN_TOKEN: a DigitalOcean API token.
//   - DIGITALOCEAN_EVAL_RUN_UUID: UUID of the evaluation run to cancel.
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

	out, _, err := client.GradientAI.CancelModelEvaluationRun(ctx, runUUID)
	if err != nil {
		panic(err)
	}

	if out.Run == nil {
		fmt.Printf("Cancel accepted for run %s (no run summary returned)\n", runUUID)
		return
	}

	run := out.Run
	fmt.Printf("Cancel accepted for run %s\n", run.EvalRunUuid)
	fmt.Printf("  name:             %s\n", run.Name)
	fmt.Printf("  status:           %s\n", run.Status)
	fmt.Printf("  candidate model:  %s (%s, source=%s)\n", run.CandidateModelName, run.CandidateModelUuid, run.CandidateModelSource)
	fmt.Printf("  judge model:      %s (%s)\n", run.JudgeModelName, run.JudgeModelUuid)
	fmt.Printf("  dataset:          %s (%s)\n", run.DatasetName, run.DatasetUuid)
	if run.CreatedAt != nil {
		fmt.Printf("  created at:       %s\n", run.CreatedAt)
	}
}
