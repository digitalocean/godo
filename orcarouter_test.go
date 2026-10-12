package godo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewOrcaRouterClient_Validation(t *testing.T) {
	if _, err := NewOrcaRouterClient(""); err == nil {
		t.Fatal("NewOrcaRouterClient with empty key: want error, got nil")
	}
	if _, err := NewOrcaRouterClient("   "); err == nil {
		t.Fatal("NewOrcaRouterClient with blank key: want error, got nil")
	}
}

func TestOrcaRouterClient_ChatCompletionsAndModels(t *testing.T) {
	const (
		apiKey = "orcarouter_api_key_xyz"
		model  = "orcarouter/fusion"
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+apiKey {
			http.Error(w, fmt.Sprintf("bad auth header %q", got), http.StatusUnauthorized)
			return
		}

		var body ChatCompletionNewParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if body.Model != model {
			http.Error(w, fmt.Sprintf("bad model %q", body.Model), http.StatusBadRequest)
			return
		}

		if body.Stream != nil && *body.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "data: %s\n\n", `{"id":"c1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"}}]}`)
			fmt.Fprintf(w, "data: %s\n\n", `{"id":"c1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":" world"}}]}`)
			fmt.Fprintf(w, "data: [DONE]\n\n")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChatCompletion{
			ID:     "chatcmpl-orc-123",
			Object: "chat.completion",
			Model:  model,
			Choices: []ChatCompletionChoice{
				{
					Index:        0,
					FinishReason: "stop",
					Message: ChatCompletionMessage{
						Role:    "assistant",
						Content: PtrTo("Hello from OrcaRouter"),
					},
				},
			},
			Usage: &ChatCompletionUsage{
				PromptTokens:     1,
				CompletionTokens: 2,
				TotalTokens:      3,
			},
		})
	})

	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+apiKey {
			http.Error(w, fmt.Sprintf("bad auth header %q", got), http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ModelList{
			Object: "list",
			Data: []InferenceModel{
				{ID: "orcarouter/fusion", Object: "model", OwnedBy: "orcarouter"},
			},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := NewOrcaRouterClient(apiKey, SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewOrcaRouterClient: %v", err)
	}

	t.Run("non-streaming chat completion", func(t *testing.T) {
		resp, _, err := client.Chat.Completions.New(context.Background(), &ChatCompletionNewParams{
			Model: model,
			Messages: []ChatCompletionMessage{
				UserMessage("Say hi"),
			},
		})
		if err != nil {
			t.Fatalf("Chat.Completions.New: %v", err)
		}
		if resp.ID != "chatcmpl-orc-123" {
			t.Fatalf("ID = %q, want chatcmpl-orc-123", resp.ID)
		}
		if got := resp.Choices[0].Message.Content; got == nil || *got != "Hello from OrcaRouter" {
			t.Fatalf("Content = %v, want %q", got, "Hello from OrcaRouter")
		}
	})

	t.Run("streaming chat completion", func(t *testing.T) {
		stream, _, err := client.Chat.Completions.NewStreaming(context.Background(), &ChatCompletionNewParams{
			Model: model,
			Messages: []ChatCompletionMessage{
				UserMessage("Say hi"),
			},
		})
		if err != nil {
			t.Fatalf("Chat.Completions.NewStreaming: %v", err)
		}
		defer stream.Close()

		var got strings.Builder
		for stream.Next() {
			ev := stream.Current()
			if len(ev.Choices) > 0 {
				got.WriteString(ev.Choices[0].Delta.Content)
			}
		}
		if err := stream.Err(); err != nil {
			t.Fatalf("stream: %v", err)
		}
		if got.String() != "Hello world" {
			t.Fatalf("streamed text = %q, want %q", got.String(), "Hello world")
		}
	})

	t.Run("list models", func(t *testing.T) {
		models, _, err := client.Models.List(context.Background())
		if err != nil {
			t.Fatalf("Models.List: %v", err)
		}
		if len(models.Data) != 1 || models.Data[0].ID != "orcarouter/fusion" {
			t.Fatalf("unexpected models: %+v", models.Data)
		}
	})
}
