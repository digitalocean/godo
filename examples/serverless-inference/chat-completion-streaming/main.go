// Command chat-completion-streaming streams a chat completion via the Serverless Inference API.
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

	stream, _, err := client.Chat.Completions.NewStreaming(ctx, &godo.ChatCompletionNewParams{
		Messages: []godo.ChatCompletionMessage{
			godo.UserMessage(question),
		},
		Model: "llama3.3-70b-instruct",
	})
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	for stream.Next() {
		evt := stream.Current()
		if len(evt.Choices) > 0 {
			print(evt.Choices[0].Delta.Content)
		}
	}
	println()

	if err := stream.Err(); err != nil {
		panic(err.Error())
	}
}
