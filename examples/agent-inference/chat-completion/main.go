// Command chat-completion runs a non-streaming chat completion against a
// DigitalOcean Agent Inference endpoint.
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

	question := "What is the capital of Portugal?"

	print("> ")
	println(question)
	println()

	completion, _, err := agent.Chat.Completions.New(ctx, &godo.ChatCompletionNewParams{
		Model: model,
		Messages: []godo.ChatCompletionMessage{
			godo.UserMessage(question),
		},
	})
	if err != nil {
		panic(err)
	}

	if msg := completion.Choices[0].Message.Content; msg != nil {
		println(*msg)
	}
}
