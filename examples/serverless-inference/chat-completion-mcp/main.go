// Command chat-completion-mcp runs a chat completion with the MCP server-side tool.
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

	completion, _, err := client.Chat.Completions.New(ctx, &godo.ChatCompletionNewParams{
		Model: "openai-gpt-oss-20b",
		Messages: []godo.ChatCompletionMessage{
			godo.UserMessage("Fetch my DigitalOcean account information and summarize it in 2 bullets."),
		},
		Tools: []godo.ChatCompletionTool{
			godo.ChatMCPTool(godo.MCPToolParams{
				ServerLabel:   "digitalocean",
				ServerURL:     "https://accounts.mcp.digitalocean.com/mcp",
				Authorization: "Bearer " + os.Getenv("DIGITALOCEAN_API_TOKEN"),
				AllowedTools:  []string{"account-get-information"},
			}),
		},
		ToolChoice: json.RawMessage(`"required"`),
		MaxTokens:  godo.PtrTo(512),
	})
	if err != nil {
		panic(err)
	}

	if msg := completion.Choices[0].Message.Content; msg != nil {
		println(msg.Text())
	}
}
