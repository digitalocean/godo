package godo

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestResponsesResponse_ToolChoicePolymorphism(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "string",
			body: `{"id":"r1","object":"response","tool_choice":"auto","usage":{"input_tokens":0,"output_tokens":0,"total_tokens":0,"input_tokens_details":{},"output_tokens_details":{}}}`,
			want: `"auto"`,
		},
		{
			name: "object",
			body: `{"id":"r2","object":"response","tool_choice":{"type":"function","name":"lookup"},"usage":{"input_tokens":0,"output_tokens":0,"total_tokens":0,"input_tokens_details":{},"output_tokens_details":{}}}`,
			want: `{"type":"function","name":"lookup"}`,
		},
		{
			name: "absent",
			body: `{"id":"r3","object":"response","usage":{"input_tokens":0,"output_tokens":0,"total_tokens":0,"input_tokens_details":{},"output_tokens_details":{}}}`,
			want: ``,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var got ResponsesResponse
			if err := json.Unmarshal([]byte(tc.body), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !bytes.Equal(got.ToolChoice, []byte(tc.want)) {
				t.Fatalf("ToolChoice = %q, want %q", string(got.ToolChoice), tc.want)
			}
		})
	}
}

// TestResponseStreamEvent_CompletedEventCarriesResponse exercises the exact
// shape that surfaced the original panic: a streamed event whose payload
// nests a ResponsesResponse with an object-shaped tool_choice. Decoding a
// real response.completed frame must succeed end-to-end via the typed
// ResponseStream event struct, not just the bare ResponsesResponse.
func TestResponseStreamEvent_CompletedEventCarriesResponse(t *testing.T) {
	const completed = `{
		"type": "response.completed",
		"sequence_number": 42,
		"response": {
			"id": "resp_123",
			"object": "response",
			"model": "openai-gpt-oss-20b",
			"tool_choice": {"type": "function", "name": "lookup"},
			"output": [],
			"usage": {
				"input_tokens": 1,
				"output_tokens": 2,
				"total_tokens": 3,
				"input_tokens_details": {},
				"output_tokens_details": {}
			}
		}
	}`

	var ev ResponseStreamEvent
	if err := json.Unmarshal([]byte(completed), &ev); err != nil {
		t.Fatalf("unmarshal completed event: %v", err)
	}
	if ev.Type != "response.completed" {
		t.Fatalf("Type = %q, want response.completed", ev.Type)
	}
	if ev.Response == nil {
		t.Fatal("Response = nil, want non-nil")
	}
	if want := `{"type": "function", "name": "lookup"}`; string(ev.Response.ToolChoice) != want {
		t.Fatalf("Response.ToolChoice = %q, want %q", string(ev.Response.ToolChoice), want)
	}
}
