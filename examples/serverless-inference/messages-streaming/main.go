// Command messages-streaming streams an Anthropic-style message from the
// Serverless Inference /v1/messages endpoint.
package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))

	ctx := context.Background()

	model := os.Getenv("DIGITALOCEAN_INFERENCE_MESSAGES_MODEL")
	if model == "" {
		model = "anthropic-claude-haiku-4.5"
	}

	content, _ := json.Marshal("Write a haiku about the ocean.")

	stream, _, err := client.Messages.NewStreaming(ctx, &godo.MessageNewParams{
		Model:     model,
		MaxTokens: 1024,
		Messages: []godo.MessageParam{
			{Role: "user", Content: content},
		},
	})
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	for stream.Next() {
		ev := stream.Current()
		if ev.Type == "content_block_delta" && ev.Delta.Type == "text_delta" {
			print(ev.Delta.Text)
		}
	}
	println()

	if err := stream.Err(); err != nil {
		panic(err)
	}
}
