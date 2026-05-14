// Command responses-streaming streams a /v1/responses call via the Serverless Inference API.
package main

import (
	"context"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))
	ctx := context.Background()

	question := "Tell me about briefly about Doug Engelbart"

	stream, _, err := client.Responses.NewStreaming(ctx, &godo.ResponseNewParams{
		Input: question,
		Model: "openai-gpt-oss-20b",
	})
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	var completeText string

	for stream.Next() {
		data := stream.Current()
		print(data.Delta)
		if data.Text != "" {
			println()
			println("Finished Content")
			completeText = data.Text
			break
		}
	}

	if stream.Err() != nil {
		panic(stream.Err())
	}

	_ = completeText
}
