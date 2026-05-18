// Command list-models lists the models reachable through the Serverless Inference API.
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

	page, _, err := client.Models.List(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d models:\n", len(page.Data))
	for _, m := range page.Data {
		fmt.Printf("  - %s  (owned_by=%s)\n", m.ID, m.OwnedBy)
	}
}
