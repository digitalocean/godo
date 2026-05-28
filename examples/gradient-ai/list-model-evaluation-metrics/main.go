// Command list-model-evaluation-metrics lists all available metrics that can
// be selected when creating a model evaluation run, via the GradientAI API.
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

	out, _, err := client.GradientAI.ListModelEvaluationMetrics(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d model evaluation metrics\n", len(out.Metrics))
	for _, metric := range out.Metrics {
		fmt.Printf("- %s (%s) type=%s value=%s category=%s inverted=%t\n",
			metric.MetricUUID, metric.MetricName, metric.MetricType, metric.MetricValueType, metric.Category, metric.Inverted)
	}
}
