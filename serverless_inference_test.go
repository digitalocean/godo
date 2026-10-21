package godo

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func assertJSONEqual(t *testing.T, got []byte, want string) {
	t.Helper()
	var gotVal, wantVal interface{}
	if err := json.Unmarshal(got, &gotVal); err != nil {
		t.Fatalf("unmarshal got: %v\n%s", err, got)
	}
	if err := json.Unmarshal([]byte(want), &wantVal); err != nil {
		t.Fatalf("unmarshal want: %v\n%s", err, want)
	}
	gotNorm, err := json.Marshal(gotVal)
	if err != nil {
		t.Fatalf("marshal got: %v", err)
	}
	wantNorm, err := json.Marshal(wantVal)
	if err != nil {
		t.Fatalf("marshal want: %v", err)
	}
	if !bytes.Equal(gotNorm, wantNorm) {
		t.Fatalf("JSON mismatch\ngot:  %s\nwant: %s", gotNorm, wantNorm)
	}
}

func TestResponseTool_ServerSideToolsMarshal(t *testing.T) {
	cases := []struct {
		name string
		tool ResponseTool
		want string
	}{
		{
			name: "web_search",
			tool: WebSearchTool(PtrTo(3), PtrTo(5)),
			want: `{"type":"web_search","max_uses":3,"max_results":5}`,
		},
		{
			name: "web_fetch",
			tool: WebFetchTool(PtrTo(3)),
			want: `{"type":"web_fetch","max_uses":3}`,
		},
		{
			name: "mcp",
			tool: MCPTool(MCPToolParams{
				ServerLabel:   "digitalocean",
				ServerURL:     "https://accounts.mcp.digitalocean.com/mcp",
				Authorization: "Bearer $DIGITALOCEAN_API_TOKEN",
				AllowedTools:  []string{"account-get-information"},
			}),
			want: `{
				"type":"mcp",
				"server_label":"digitalocean",
				"server_url":"https://accounts.mcp.digitalocean.com/mcp",
				"authorization":"Bearer $DIGITALOCEAN_API_TOKEN",
				"allowed_tools":["account-get-information"]
			}`,
		},
		{
			name: "knowledge_base_retrieval",
			tool: KnowledgeBaseRetrievalTool("09d65da1-5225-11f1-b074-4e013e2dxxxx"),
			want: `{"type":"knowledge_base_retrieval","knowledge_base_id":"09d65da1-5225-11f1-b074-4e013e2dxxxx"}`,
		},
		{
			name: "model_synthesis_preset",
			tool: ModelSynthesisTool(ModelSynthesisToolParams{Preset: "balanced"}),
			want: `{"type":"model_synthesis","preset":"balanced"}`,
		},
		{
			name: "model_synthesis_direct",
			tool: ModelSynthesisTool(ModelSynthesisToolParams{
				AnalysisModels: []string{"openai-gpt-5.5", "deepseek-v4-pro"},
				MaxToolCalls:   PtrTo(8),
				MaxTokens:      PtrTo(16000),
			}),
			want: `{
				"type":"model_synthesis",
				"analysis_models":["openai-gpt-5.5","deepseek-v4-pro"],
				"max_tool_calls":8,
				"max_tokens":16000
			}`,
		},
		{
			name: "tool_search",
			tool: ToolSearchTool(),
			want: `{"type":"tool_search"}`,
		},
		{
			name: "function_with_defer_loading",
			tool: ResponseTool{
				Type:         "function",
				Name:         "get_weather",
				Description:  "Get weather",
				Parameters:   map[string]interface{}{"type": "object"},
				DeferLoading: PtrTo(true),
			},
			want: `{
				"type":"function",
				"name":"get_weather",
				"description":"Get weather",
				"parameters":{"type":"object"},
				"defer_loading":true
			}`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.tool)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			assertJSONEqual(t, got, tc.want)

			var roundTrip ResponseTool
			if err := json.Unmarshal(got, &roundTrip); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			again, err := json.Marshal(roundTrip)
			if err != nil {
				t.Fatalf("re-marshal: %v", err)
			}
			assertJSONEqual(t, again, tc.want)
		})
	}
}

func TestChatCompletionTool_ServerSideToolsMarshal(t *testing.T) {
	cases := []struct {
		name string
		tool ChatCompletionTool
		want string
	}{
		{
			name: "web_search",
			tool: ChatWebSearchTool(PtrTo(3), PtrTo(5)),
			want: `{"type":"web_search","max_uses":3,"max_results":5}`,
		},
		{
			name: "web_fetch",
			tool: ChatWebFetchTool(PtrTo(3)),
			want: `{"type":"web_fetch","max_uses":3}`,
		},
		{
			name: "mcp",
			tool: ChatMCPTool(MCPToolParams{
				ServerLabel:   "digitalocean",
				ServerURL:     "https://accounts.mcp.digitalocean.com/mcp",
				Authorization: "Bearer $DIGITALOCEAN_API_TOKEN",
				AllowedTools:  []string{"account-get-information"},
			}),
			want: `{
				"type":"mcp",
				"server_label":"digitalocean",
				"server_url":"https://accounts.mcp.digitalocean.com/mcp",
				"authorization":"Bearer $DIGITALOCEAN_API_TOKEN",
				"allowed_tools":["account-get-information"]
			}`,
		},
		{
			name: "knowledge_base_retrieval",
			tool: ChatKnowledgeBaseRetrievalTool("<your-knowledge-base-id>"),
			want: `{"type":"knowledge_base_retrieval","knowledge_base_id":"<your-knowledge-base-id>"}`,
		},
		{
			name: "model_synthesis",
			tool: ChatModelSynthesisTool(ModelSynthesisToolParams{
				AnalysisModels: []string{"openai-gpt-5.5", "deepseek-v4-pro"},
				MaxToolCalls:   PtrTo(8),
				MaxTokens:      PtrTo(16000),
			}),
			want: `{
				"type":"model_synthesis",
				"analysis_models":["openai-gpt-5.5","deepseek-v4-pro"],
				"max_tool_calls":8,
				"max_tokens":16000
			}`,
		},
		{
			name: "function",
			tool: ChatFunctionTool("lookup", "Look something up", map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"q": map[string]interface{}{"type": "string"},
				},
			}),
			want: `{
				"type":"function",
				"function":{
					"name":"lookup",
					"description":"Look something up",
					"parameters":{"type":"object","properties":{"q":{"type":"string"}}}
				}
			}`,
		},
		{
			name: "tool_search",
			tool: ChatToolSearchTool(),
			want: `{"type":"tool_search"}`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.tool)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			assertJSONEqual(t, got, tc.want)

			var roundTrip ChatCompletionTool
			if err := json.Unmarshal(got, &roundTrip); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			again, err := json.Marshal(roundTrip)
			if err != nil {
				t.Fatalf("re-marshal: %v", err)
			}
			assertJSONEqual(t, again, tc.want)
		})
	}
}

func TestChatCompletionMessageContent_StringAndParts(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		msg := UserMessage("Describe this image.")
		got, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		assertJSONEqual(t, got, `{"role":"user","content":"Describe this image."}`)

		var roundTrip ChatCompletionMessage
		if err := json.Unmarshal(got, &roundTrip); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if roundTrip.Content.Text() != "Describe this image." {
			t.Fatalf("Text() = %q", roundTrip.Content.Text())
		}
	})

	t.Run("multimodal", func(t *testing.T) {
		msg := UserMessageParts(
			TextPart("Describe this image."),
			ImageURLPart("data:image/jpeg;base64,/9j/4AAQ"),
		)
		got, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		assertJSONEqual(t, got, `{
			"role":"user",
			"content":[
				{"type":"text","text":"Describe this image."},
				{"type":"image_url","image_url":{"url":"data:image/jpeg;base64,/9j/4AAQ"}}
			]
		}`)

		var roundTrip ChatCompletionMessage
		if err := json.Unmarshal(got, &roundTrip); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if roundTrip.Content.Text() != "Describe this image." {
			t.Fatalf("Text() = %q", roundTrip.Content.Text())
		}
		parts := roundTrip.Content.Parts()
		if len(parts) != 2 || parts[1].ImageURL == nil || parts[1].ImageURL.URL != "data:image/jpeg;base64,/9j/4AAQ" {
			t.Fatalf("Parts() = %+v", parts)
		}
	})
}

func TestSpeechNewParams_Marshal(t *testing.T) {
	params := SpeechNewParams{
		Model:          "qwen3-tts-voicedesign",
		Input:          "Hello, world!",
		Voice:          "alloy",
		ResponseFormat: "mp3",
		Instructions:   "Speak naturally.",
	}
	got, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	assertJSONEqual(t, got, `{
		"model":"qwen3-tts-voicedesign",
		"input":"Hello, world!",
		"voice":"alloy",
		"response_format":"mp3",
		"instructions":"Speak naturally."
	}`)
}

func TestSpeechResponse_Unmarshal(t *testing.T) {
	const body = `{"created":1710700000,"data":[{"b64_json":"YWJj"}]}`
	var got SpeechResponse
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Created != 1710700000 || len(got.Data) != 1 || got.Data[0].B64JSON != "YWJj" {
		t.Fatalf("SpeechResponse = %+v", got)
	}
}

func TestVideoGenerateParams_Marshal(t *testing.T) {
	params := VideoGenerateParams{
		Model:  "wan2-2-t2v-a14b",
		Prompt: "Two cats boxing on a spotlighted stage",
		Size:   "1280x720",
		FPS:    PtrTo(16),
	}
	got, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	assertJSONEqual(t, got, `{
		"model":"wan2-2-t2v-a14b",
		"prompt":"Two cats boxing on a spotlighted stage",
		"size":"1280x720",
		"fps":16
	}`)
}

func TestVideoGeneration_Unmarshal(t *testing.T) {
	const body = `{
		"created_at": 1777003604,
		"error": null,
		"id": "video_abc",
		"model": "wan2-2-t2v-a14b",
		"object": "video",
		"output": {"url":"https://example.com/out.mp4"},
		"status": "completed",
		"x_request_id": null
	}`
	var got VideoGeneration
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != "video_abc" || got.Status != "completed" || got.Model != "wan2-2-t2v-a14b" {
		t.Fatalf("VideoGeneration = %+v", got)
	}
	if got.Output == nil || got.Output.URL != "https://example.com/out.mp4" {
		t.Fatalf("Output = %+v", got.Output)
	}
}

func TestSpeechService_New_Binary(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/audio/speech", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		var body SpeechNewParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Model != "qwen3-tts-voicedesign" || body.Input != "hi" || body.ResponseFormat != "wav" {
			t.Fatalf("body = %+v", body)
		}
		w.Header().Set("Content-Type", "audio/wav")
		_, _ = w.Write([]byte("RIFF....WAVE"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewClient(nil)
	base, _ := url.Parse(srv.URL + "/")
	client.Audio.Speech.baseURL = base

	got, _, err := client.Audio.Speech.New(context.Background(), &SpeechNewParams{
		Model:          "qwen3-tts-voicedesign",
		Input:          "hi",
		ResponseFormat: "wav",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	audio, err := got.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if string(audio) != "RIFF....WAVE" || got.ContentType != "audio/wav" {
		t.Fatalf("got = %+v audio=%q", got, audio)
	}
}

func TestSpeechService_New_JSONEnvelope(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/audio/speech", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"created":1,"data":[{"b64_json":"AQID"}]}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewClient(nil)
	base, _ := url.Parse(srv.URL + "/")
	client.Audio.Speech.baseURL = base

	got, _, err := client.Audio.Speech.New(context.Background(), &SpeechNewParams{
		Model: "qwen3-tts-voicedesign",
		Input: "hi",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	audio, err := got.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if string(audio) != "\x01\x02\x03" {
		t.Fatalf("audio = %q", audio)
	}
}

func TestVideoGenerationService_CreateGetContent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/videos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		var body VideoGenerateParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Model != "wan2-2-t2v-a14b" || body.Prompt == "" || body.Size != "1280x720" {
			t.Fatalf("body = %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"video_abc123","status":"queued"}`)
	})
	mux.HandleFunc("/v1/videos/video_abc123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"video_abc123","status":"completed","model":"wan2-2-t2v-a14b","output":null}`)
	})
	mux.HandleFunc("/v1/videos/video_abc123/content", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = w.Write([]byte("mp4-bytes"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewClient(nil)
	base, _ := url.Parse(srv.URL + "/")
	client.Video.Generations.baseURL = base

	job, _, err := client.Video.Generations.Create(context.Background(), &VideoGenerateParams{
		Model:  "wan2-2-t2v-a14b",
		Prompt: "cats boxing",
		Size:   "1280x720",
		FPS:    PtrTo(16),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if job.ID != "video_abc123" || job.Status != "queued" {
		t.Fatalf("Create job = %+v", job)
	}

	done, _, err := client.Video.Generations.Get(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if done.ID != "video_abc123" || done.Status != "completed" {
		t.Fatalf("Get = %+v", done)
	}

	raw, _, err := client.Video.Generations.Content(context.Background(), done.ID)
	if err != nil {
		t.Fatalf("Content: %v", err)
	}
	if string(raw) != "mp4-bytes" {
		t.Fatalf("Content = %q", raw)
	}
}
