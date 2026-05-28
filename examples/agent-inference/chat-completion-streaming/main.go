package main

import (
	"context"
	"fmt"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	baseURL := os.Getenv("DIGITALOCEAN_AGENT_URL")
	accessKey := os.Getenv("DIGITALOCEAN_AGENT_ACCESS_KEY")
	if baseURL == "" || accessKey == "" {
		fmt.Fprintln(os.Stderr, "DIGITALOCEAN_AGENT_URL and DIGITALOCEAN_AGENT_ACCESS_KEY must be set")
		os.Exit(2)
	}

	agent, err := godo.NewAgentInferenceClient(baseURL, accessKey)
	if err != nil {
		panic(err)
	}

	model := os.Getenv("DIGITALOCEAN_AGENT_MODEL")
	if model == "" {
		model = "llama3.3-70b-instruct"
	}

	ctx := context.Background()

	question := "Write me a haiku"

	print("> ")
	println(question)
	println()

	stream, _, err := agent.Chat.Completions.NewStreaming(ctx, &godo.ChatCompletionNewParams{
		Model: model,
		Messages: []godo.ChatCompletionMessage{
			godo.UserMessage(question),
		},
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
