// Command messages sends a non-streaming message via the Anthropic-compatible
// /v1/messages endpoint of the Serverless Inference API.
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
		model = "claude-opus-4-6"
	}

	content, _ := json.Marshal("What is the capital of Portugal?")

	msg, _, err := client.Messages.New(ctx, &godo.MessageNewParams{
		Model:     model,
		MaxTokens: 1024,
		Messages: []godo.MessageParam{
			{Role: "user", Content: content},
		},
	})
	if err != nil {
		panic(err)
	}

	for _, block := range msg.Content {
		if block.Type == "text" {
			println(block.Text)
		}
	}
}
