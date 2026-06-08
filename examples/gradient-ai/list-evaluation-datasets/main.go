// Command list-evaluation-datasets lists evaluation datasets via the GradientAI
// API. Results can optionally be filtered by dataset type.
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

	out, _, err := client.GradientAI.ListEvaluationDatasets(ctx, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d evaluation datasets\n", len(out.EvaluationDatasets))
	for _, dataset := range out.EvaluationDatasets {
		fmt.Printf("- %s (%s) type=%s rows=%d ground_truth=%t\n",
			dataset.DatasetUUID, dataset.DatasetName, dataset.DatasetType, dataset.RowCount, dataset.HasGroundTruth)
	}
}
