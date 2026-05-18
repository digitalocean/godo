// Command chat-completion runs a non-streaming chat completion via the Serverless Inference API.
package main

import (
	"context"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))

	ctx := context.Background()

	question := "Write me a haiku"

	print("> ")
	println(question)
	println()

	completion, _, err := client.Chat.Completions.New(ctx, &godo.ChatCompletionNewParams{
		Messages: []godo.ChatCompletionMessage{
			godo.UserMessage(question),
		},
		Model: "llama3.3-70b-instruct",
	})
	if err != nil {
		panic(err)
	}

	if msg := completion.Choices[0].Message.Content; msg != nil {
		println(*msg)
	}
}
