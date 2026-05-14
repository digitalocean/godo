package godo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAgentInferenceClient_NewAndStreaming(t *testing.T) {
	const (
		accessKey = "agent_access_key_xyz"
		model     = "llama3.3-70b-instruct"
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if got := r.URL.Query().Get("agent"); got != "true" {
			http.Error(w, fmt.Sprintf("missing ?agent=true (got %q)", got), http.StatusBadRequest)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+accessKey {
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
			// Streaming branch: emit two delta chunks then [DONE].
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "data: %s\n\n", `{"id":"c1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"}}]}`)
			fmt.Fprintf(w, "data: %s\n\n", `{"id":"c1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":" world"}}]}`)
			fmt.Fprintf(w, "data: [DONE]\n\n")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChatCompletion{
			ID:     "chatcmpl-abc123",
			Object: "chat.completion",
			Model:  model,
			Choices: []ChatCompletionChoice{
				{
					Index:        0,
					FinishReason: "stop",
					Message: ChatCompletionMessage{
						Role:    "assistant",
						Content: PtrTo("Hello world"),
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

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := NewAgentInferenceClient(srv.URL, accessKey)
	if err != nil {
		t.Fatalf("NewAgentInferenceClient: %v", err)
	}

	t.Run("non-streaming", func(t *testing.T) {
		resp, _, err := client.Chat.Completions.New(context.Background(), &ChatCompletionNewParams{
			Model: model,
			Messages: []ChatCompletionMessage{
				UserMessage("Say hi"),
			},
		})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if resp.ID != "chatcmpl-abc123" {
			t.Fatalf("ID = %q, want chatcmpl-abc123", resp.ID)
		}
		if got := resp.Choices[0].Message.Content; got == nil || *got != "Hello world" {
			t.Fatalf("Content = %v, want %q", got, "Hello world")
		}
	})

	t.Run("streaming", func(t *testing.T) {
		stream, _, err := client.Chat.Completions.NewStreaming(context.Background(), &ChatCompletionNewParams{
			Model: model,
			Messages: []ChatCompletionMessage{
				UserMessage("Say hi"),
			},
		})
		if err != nil {
			t.Fatalf("NewStreaming: %v", err)
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
			t.Fatalf("stream.Err: %v", err)
		}
		if got.String() != "Hello world" {
			t.Fatalf("concatenated stream = %q, want %q", got.String(), "Hello world")
		}
	})
}

// TestAgentInferenceClient_ErrorResponse verifies that a non-2xx response is
// returned as an *ErrorResponse (matching godo.Client.Do's behaviour).
func TestAgentInferenceClient_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"id":"forbidden","message":"this model is not available for your subscription tier","request_id":"abc"}`)
	}))
	defer srv.Close()

	client, err := NewAgentInferenceClient(srv.URL, "agent_access_key_xyz")
	if err != nil {
		t.Fatalf("NewAgentInferenceClient: %v", err)
	}

	_, resp, err := client.Chat.Completions.New(context.Background(), &ChatCompletionNewParams{
		Model:    "anything",
		Messages: []ChatCompletionMessage{UserMessage("hi")},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 response, got %+v", resp)
	}
	var er *ErrorResponse
	if !errors.As(err, &er) {
		t.Fatalf("expected *ErrorResponse, got %T (%v)", err, err)
	}
	if !strings.Contains(er.Message, "subscription tier") {
		t.Fatalf("Message = %q, want substring %q", er.Message, "subscription tier")
	}
}

// TestNewAgentInferenceClient_Validation pins constructor argument checks.
func TestNewAgentInferenceClient_Validation(t *testing.T) {
	cases := []struct {
		name      string
		baseURL   string
		accessKey string
		wantErr   string
	}{
		{"empty url", "", "k", "baseURL is required"},
		{"empty key", "https://x.agents.do-ai.run", "", "accessKey is required"},
		{"bad scheme", "::::not-a-url", "k", "parse baseURL"},
		{"missing scheme", "x.agents.do-ai.run", "k", "missing scheme or host"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewAgentInferenceClient(tc.baseURL, tc.accessKey)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tc.wantErr)
			}
		})
	}
}
