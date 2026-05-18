// Command embeddings creates embedding vectors via the Serverless Inference API.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))

	timeout := 60 * time.Second
	if v := os.Getenv("DIGITALOCEAN_INFERENCE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			timeout = d
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	model := os.Getenv("DIGITALOCEAN_INFERENCE_EMBEDDINGS_MODEL")
	if model == "" {
		model = "qwen3-embedding-0.6b"
	}

	inputs := []string{
		"The quick brown fox jumps over the lazy dog.",
		"DigitalOcean's Serverless Inference API speaks the OpenAI dialect.",
		"Embedding vectors are useful for similarity search and RAG.",
	}

	fmt.Fprintf(os.Stderr, "POST /v1/embeddings  model=%s  inputs=%d  timeout=%s ...\n",
		model, len(inputs), timeout)
	start := time.Now()

	resp, _, err := client.Embeddings.New(ctx, &godo.EmbeddingNewParams{
		Model: model,
		Input: inputs,
	})
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stderr, "received response in %s\n", time.Since(start).Round(time.Millisecond))

	fmt.Printf("model=%s  vectors=%d\n", resp.Model, len(resp.Data))
	for i, e := range resp.Data {
		var vec []float32
		if err := json.Unmarshal(e.Embedding, &vec); err != nil {
			panic(fmt.Errorf("decode vector %d: %w", i, err))
		}
		head := vec
		if len(head) > 5 {
			head = head[:5]
		}
		fmt.Printf("  [%d] dim=%d  head=%v...  input=%q\n", e.Index, len(vec), head, inputs[i])
	}
	fmt.Printf("usage: prompt_tokens=%d total_tokens=%d\n",
		resp.Usage.PromptTokens, resp.Usage.TotalTokens)
}
