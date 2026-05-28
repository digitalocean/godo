// Command responses runs a non-streaming /v1/responses call via the Serverless Inference API.
package main

import (
	"context"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))
	ctx := context.Background()

	question := "Write me a haiku about computers"

	resp, _, err := client.Responses.New(ctx, &godo.ResponseNewParams{
		Input: question,
		Model: "openai-gpt-oss-20b",
	})

	if err != nil {
		panic(err)
	}

	println(resp.OutputText())
}
