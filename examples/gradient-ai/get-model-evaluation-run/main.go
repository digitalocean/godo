// Command get-model-evaluation-run retrieves a model evaluation run, including
// the run summary, aggregated result metrics, and a paginated list of
// per-prompt results, via the GradientAI API.
//
// Required env vars:
//   - DIGITALOCEAN_TOKEN: a DigitalOcean API token.
//   - DIGITALOCEAN_EVAL_RUN_UUID: UUID of the evaluation run to retrieve.
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

	out, _, err := client.GradientAI.GetModelEvaluationRun(ctx, runUUID, &godo.ModelEvaluationRunGetOptions{
		Page:    1,
		PerPage: 50,
	})
	if err != nil {
		panic(err)
	}

	if out.Run == nil {
		fmt.Printf("No run detail returned for %s\n", runUUID)
		return
	}

	run := out.Run
	fmt.Printf("Model evaluation run %s\n", run.EvalRunUuid)
	fmt.Printf("  name:             %s\n", run.Name)
	fmt.Printf("  status:           %s\n", run.Status)
	fmt.Printf("  candidate model:  %s (%s, source=%s)\n", run.CandidateModelName, run.CandidateModelUuid, run.CandidateModelSource)
	fmt.Printf("  judge model:      %s (%s)\n", run.JudgeModelName, run.JudgeModelUuid)
	fmt.Printf("  dataset:          %s (%s)\n", run.DatasetName, run.DatasetUuid)

	if rs := run.ResultSummary; rs != nil {
		fmt.Printf("  overall score:    %.2f%%\n", rs.OverallScorePercent)
		fmt.Printf("  duration:         %ds\n", rs.TotalDurationSeconds)
		for _, m := range rs.MetricSummaries {
			fmt.Printf("    metric %s: pass=%.1f%% fail=%.1f%%\n", m.MetricName, m.PassPercent, m.FailPercent)
		}
	}

	fmt.Printf("  results returned: %d\n", len(out.Results))
}
