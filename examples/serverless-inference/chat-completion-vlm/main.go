// Command chat-completion-vlm runs a multimodal chat completion with text and image content.
package main

import (
	"context"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))
	ctx := context.Background()

	// IMAGE_URL may be an HTTPS URL or a data:image/...;base64,... URI.
	imageURL := os.Getenv("IMAGE_URL")
	if imageURL == "" {
		imageURL = "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3b/Online_Bicycles.jpg/640px-Online_Bicycles.jpg"
	}

	completion, _, err := client.Chat.Completions.New(ctx, &godo.ChatCompletionNewParams{
		Model: "nemotron-nano-12b-v2-vl",
		Messages: []godo.ChatCompletionMessage{
			godo.UserMessageParts(
				godo.TextPart("Describe this image."),
				godo.ImageURLPart(imageURL),
			),
		},
		MaxTokens: godo.PtrTo(512),
	})
	if err != nil {
		panic(err)
	}

	if msg := completion.Choices[0].Message.Content; msg != nil {
		println(msg.Text())
	}
}
