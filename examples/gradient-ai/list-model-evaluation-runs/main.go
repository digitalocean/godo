// Command list-model-evaluation-runs lists model evaluation runs, optionally
// filtered by preset UUID, status (single or multiple), candidate model source
// types, and free-text search. Results can also be sorted, via the GradientAI
// API.
//
// Required env vars:
//   - DIGITALOCEAN_TOKEN: a DigitalOcean API token.
//
// Optional env vars:
//   - DIGITALOCEAN_EVAL_PRESET_UUID: filter results to runs from this preset.
//   - DIGITALOCEAN_EVAL_RUN_STATUS: filter by a single run status (e.g.
//     MODEL_EVALUATION_RUN_SUCCESSFUL).
//   - DIGITALOCEAN_EVAL_RUN_STATUSES: comma-separated list of run statuses to
//     include.
//   - DIGITALOCEAN_EVAL_CANDIDATE_TYPES: comma-separated list of candidate
//     model source types (e.g. CANDIDATE_MODEL_SOURCE_SERVERLESS).
//   - DIGITALOCEAN_EVAL_SEARCH: free-text substring to search across run,
//     candidate model, and dataset names.
//   - DIGITALOCEAN_EVAL_SORT_BY: sort field (e.g.
//     MODEL_EVALUATION_RUN_SORT_FIELD_CREATED_AT).
//   - DIGITALOCEAN_EVAL_SORT_DIRECTION: sort direction (SORT_DIRECTION_ASC or
//     SORT_DIRECTION_DESC).
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))

	opt := &godo.ModelEvaluationRunListOptions{
		EvalPresetUUID: os.Getenv("DIGITALOCEAN_EVAL_PRESET_UUID"),
		Status:         godo.ModelEvaluationRunStatus(os.Getenv("DIGITALOCEAN_EVAL_RUN_STATUS")),
		Search:         os.Getenv("DIGITALOCEAN_EVAL_SEARCH"),
		SortBy:         godo.ModelEvaluationRunSortField(os.Getenv("DIGITALOCEAN_EVAL_SORT_BY")),
		SortDirection:  godo.ModelEvaluationRunSortDirection(os.Getenv("DIGITALOCEAN_EVAL_SORT_DIRECTION")),
	}

	for _, s := range splitCSV(os.Getenv("DIGITALOCEAN_EVAL_RUN_STATUSES")) {
		opt.Statuses = append(opt.Statuses, godo.ModelEvaluationRunStatus(s))
	}
	for _, c := range splitCSV(os.Getenv("DIGITALOCEAN_EVAL_CANDIDATE_TYPES")) {
		opt.CandidateTypes = append(opt.CandidateTypes, godo.CandidateModelSource(c))
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

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
