// Command responses-web-search runs a /v1/responses call with the web_search server-side tool.
package main

import (
	"context"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))
	ctx := context.Background()

	resp, _, err := client.Responses.New(ctx, &godo.ResponseNewParams{
		Model: "openai-gpt-oss-20b",
		Input: "What are the latest pricing changes for DigitalOcean Droplets?",
		Tools: []godo.ResponseTool{
			godo.WebSearchTool(godo.PtrTo(3), godo.PtrTo(5)),
		},
		MaxOutputTokens: godo.PtrTo(1024),
	})
	if err != nil {
		panic(err)
	}

	println(resp.OutputText())
}
