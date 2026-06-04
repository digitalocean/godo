// Command create-model-evaluation-run creates a new model evaluation run via
// the GradientAI API.
//
// Required env vars:
//   - DIGITALOCEAN_TOKEN: a DigitalOcean API token.
//   - DIGITALOCEAN_EVAL_RUN_NAME: human-readable name for the run.
//   - DIGITALOCEAN_CANDIDATE_MODEL_UUID: UUID of the candidate model to evaluate.
//   - DIGITALOCEAN_JUDGE_MODEL_UUID: UUID of the judge model used to score responses.
//   - DIGITALOCEAN_EVAL_DATASET_UUID: UUID of the dataset to use for evaluation.
//   - DIGITALOCEAN_EVAL_METRIC_UUIDS: comma-separated list of metric UUIDs.
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

	required := func(name string) string {
		v := os.Getenv(name)
		if v == "" {
			fmt.Fprintf(os.Stderr, "%s is required\n", name)
			os.Exit(1)
		}
		return v
	}

	createRequest := &godo.CreateModelEvaluationRunRequest{
		Name:               required("DIGITALOCEAN_EVAL_RUN_NAME"),
		CandidateModelUUID: required("DIGITALOCEAN_CANDIDATE_MODEL_UUID"),
		JudgeModelUUID:     required("DIGITALOCEAN_JUDGE_MODEL_UUID"),
		DatasetUUID:        required("DIGITALOCEAN_EVAL_DATASET_UUID"),
		MetricUUIDs:        strings.Split(required("DIGITALOCEAN_EVAL_METRIC_UUIDS"), ","),
	}

	ctx := context.Background()

	out, _, err := client.GradientAI.CreateModelEvaluationRun(ctx, createRequest)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created model evaluation run: %s\n", out.EvalRunUuid)
}
