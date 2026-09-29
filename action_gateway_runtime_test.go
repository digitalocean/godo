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
	"time"

	"golang.org/x/oauth2"
)

func actionGatewayTestSession(t *testing.T) (*ActionGatewaySession, *httptest.Server) {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == actionGatewayPath+"/sessions" {
			if request.Method != http.MethodPost {
				t.Errorf("session create method = %s", request.Method)
			}
			var body ActionGatewaySessionCreateRequest
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body.Name == "" || body.ActorID != "alice" || body.Policy == nil || body.Policy.DefaultAction != "ask" {
				t.Errorf("session create body = %+v", body)
			}
			_, _ = fmt.Fprint(writer, `{"session":{"sessionUrn":"do:managed_agent_session:abc","actorId":"alice"},"mcpUrl":"https://actions.do-ai.run/mcp/session/abc","tools":[]}`)
			return
		}
		if request.Header.Get("X-Actor-Id") != "alice" || request.Header.Get("X-Session-Id") != "abc" || request.Header.Get("MCP-Protocol-Version") != actionGatewayMCPVersion {
			t.Errorf("missing MCP headers: %v", request.Header)
		}
		if strings.HasPrefix(request.URL.Path, "/approvals/") {
			if request.URL.EscapedPath() != "/approvals/approval%2F1" {
				t.Errorf("approval path = %s", request.URL.EscapedPath())
			}
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		if request.URL.Path != "/mcp/session/abc" {
			t.Errorf("MCP path = %s", request.URL.Path)
		}
		var envelope struct {
			ID     int64  `json:"id"`
			Method string `json:"method"`
			Params struct {
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			} `json:"params"`
		}
		if err := json.NewDecoder(request.Body).Decode(&envelope); err != nil {
			t.Error(err)
		}
		var result string
		switch envelope.Method {
		case "tools/list":
			result = `{"tools":[{"name":"action_search","inputSchema":{"anyOf":[{}],"properties":{}}},{"name":"action_invoke"},{"name":"action_code"},{"name":"fetch","description":"Fetch","inputSchema":{"type":"object","properties":{"url":{"type":"string"}}}}]}`
		case "tools/call":
			switch envelope.Params.Name {
			case actionGatewayInvokeTool:
				var arguments struct {
					Tools []ActionGatewayInvokeTool `json:"tools"`
				}
				if err := json.Unmarshal(envelope.Params.Arguments, &arguments); err != nil {
					t.Error(err)
				}
				results := make([]map[string]interface{}, len(arguments.Tools))
				for index, tool := range arguments.Tools {
					outcome := map[string]interface{}{"status": "succeeded", "output": map[string]interface{}{"tool": tool.Tool}}
					if tool.Tool == "bad" {
						outcome = map[string]interface{}{"status": "failed", "error": map[string]interface{}{"message": "permission denied", "class": "forbidden"}, "_meta": map[string]string{"approval_id": "approval-1"}}
					}
					results[index] = map[string]interface{}{"tool": tool.Tool, "result": outcome}
				}
				encoded, _ := json.Marshal(map[string]interface{}{"total_count": len(results), "results": results})
				result = `{"structuredContent":` + string(encoded) + `}`
			case actionGatewaySearchTool:
				result = `{"structuredContent":{"results":[{"results":[{"name":"fetch","description":"Fetch","inputSchema":{"type":"object"}}]}]}}`
			case actionGatewayCodeTool:
				writer.Header().Set("Content-Type", "text/event-stream")
				_, _ = fmt.Fprintf(writer, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":%d,\"result\":{\"structuredContent\":{\"stdout\":\"hi\"}}}\n\n", envelope.ID)
				return
			case "fetch":
				result = `{"structuredContent":{"ok":true}}`
			case "error":
				result = `{"isError":true,"structuredContent":{"error":{"message":"permission denied","class":"forbidden"}}}`
			case "nested_error":
				result = `{"isError":true,"structuredContent":{"error":{"message":"failed","invocation_id":"nested-id"}}}`
			case "rpc_error":
				_, _ = fmt.Fprintf(writer, `{"jsonrpc":"2.0","id":%d,"error":{"code":-32001,"message":"unavailable"}}`, envelope.ID)
				return
			default:
				t.Errorf("unexpected tool = %q", envelope.Params.Name)
			}
		default:
			t.Errorf("unexpected method = %s", envelope.Method)
		}
		_, _ = fmt.Fprintf(writer, `{"jsonrpc":"2.0","id":%d,"result":%s}`, envelope.ID, result)
	}))
	client, err := New(server.Client(), SetBaseURL(server.URL), WithActionGatewayMCPBaseURL(server.URL))
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	session, _, err := client.ActionGateway.Sessions.CreateRuntime(context.Background(), &ActionGatewaySessionCreateRequest{ActorID: "alice", Tools: []string{}})
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	return session, server
}

func TestActionGatewaySessionRuntime(t *testing.T) {
	session, server := actionGatewayTestSession(t)
	defer server.Close()
	ctx := context.Background()
	if !strings.HasPrefix(session.MCPURL, server.URL) {
		t.Fatalf("override not applied: %s", session.MCPURL)
	}
	session.MCPURL = "http://example.com/unsafe"
	session.ActorID = "other"
	session.Record.SessionURN = "urn:other"
	meta, _, err := session.Tools.List(ctx, false)
	if err != nil || len(meta) != 3 {
		t.Fatalf("meta tools = %+v, %v", meta, err)
	}
	all, _, err := session.Tools.List(ctx, true)
	if err != nil || len(all) != 4 {
		t.Fatalf("all tools = %+v, %v", all, err)
	}
	search, _, err := session.Tools.Search(ctx, &ActionGatewayToolSearchRequest{Queries: []ActionGatewayToolQuery{{UseCase: "fetch a URL"}}})
	if err != nil || !strings.Contains(string(search), "fetch") {
		t.Fatalf("search = %s, %v", search, err)
	}
	output, _, err := session.Tools.InvokeOne(ctx, "fetch", json.RawMessage(`{"url":"https://example.com"}`), "")
	if err != nil || !strings.Contains(string(output), "fetch") {
		t.Fatalf("invoke = %s, %v", output, err)
	}
	_, _, err = session.Tools.InvokeOne(ctx, "bad", json.RawMessage(`{}`), "")
	var toolError *ActionGatewayToolError
	if !errors.As(err, &toolError) || toolError.Class != "forbidden" || !strings.Contains(string(toolError.Meta), "approval-1") {
		t.Fatalf("expected tool error, got %v", err)
	}
	output, _, err = session.Tools.Call(ctx, "fetch", map[string]interface{}{"url": "example.com"})
	if err != nil || string(output) != `{"ok":true}` {
		t.Fatalf("direct call = %s, %v", output, err)
	}
	_, _, err = session.Tools.Call(ctx, "nested_error", nil)
	var nestedError *ActionGatewayToolError
	if !errors.As(err, &nestedError) || nestedError.InvocationID != "nested-id" {
		t.Fatalf("nested invocation ID = %v", err)
	}
	output, _, err = session.Code.Execute(ctx, "print('hi')", "test")
	if err != nil || !strings.Contains(string(output), "hi") {
		t.Fatalf("SSE code = %s, %v", output, err)
	}
	if _, response, err := session.Approve(ctx, "approval/1"); err != nil || response.StatusCode != http.StatusNoContent {
		t.Fatalf("approve = %v, %v", response, err)
	}
	if _, response, err := session.Deny(ctx, "approval/1"); err != nil || response.StatusCode != http.StatusNoContent {
		t.Fatalf("deny = %v, %v", response, err)
	}
	_, _, err = session.Tools.Call(ctx, "error", nil)
	if !errors.As(err, &toolError) || toolError.Message != "permission denied" {
		t.Fatalf("expected tool error, got %v", err)
	}
	_, _, err = session.Tools.Call(ctx, "rpc_error", nil)
	var protocolError *ActionGatewayProtocolError
	if !errors.As(err, &protocolError) || protocolError.Code != -32001 {
		t.Fatalf("expected protocol error, got %v", err)
	}
}

func TestActionGatewayInferenceAdapters(t *testing.T) {
	session, server := actionGatewayTestSession(t)
	defer server.Close()
	ctx := context.Background()
	chatTools, _, err := session.ChatTools(ctx, nil)
	if err != nil || len(chatTools) != 3 || chatTools[0].Function.Parameters["type"] != "object" {
		t.Fatalf("chat tools = %+v, %v", chatTools, err)
	}
	messageTools, _, err := session.MessageTools(ctx, &ActionGatewayInferenceToolsOptions{Names: []string{"fetch"}})
	if err != nil || len(messageTools) != 1 || messageTools[0].Name != "fetch" {
		t.Fatalf("message tools = %+v, %v", messageTools, err)
	}
	responseTools, _, err := session.ResponseTools(ctx, &ActionGatewayInferenceToolsOptions{Search: &ActionGatewayToolSearchRequest{Queries: []ActionGatewayToolQuery{{UseCase: "fetch"}}}})
	if err != nil || len(responseTools) != 1 || responseTools[0].Name != "fetch" {
		t.Fatalf("response tools = %+v, %v", responseTools, err)
	}
	chat := &ChatCompletion{Choices: []ChatCompletionChoice{{Message: ChatCompletionMessage{ToolCalls: []ChatCompletionToolCall{{ID: "chat-1", Function: ChatCompletionToolCallFunc{Name: "fetch", Arguments: `{}`}}, {ID: "chat-2", Function: ChatCompletionToolCallFunc{Name: "bad", Arguments: `{}`}}}}}}}
	chatResults, _, err := session.HandleChatToolCalls(ctx, chat)
	if err != nil || len(chatResults) != 2 || chatResults[0].Role != "tool" || !strings.Contains(*chatResults[1].Content, "permission denied") || !strings.Contains(*chatResults[1].Content, "approval-1") {
		t.Fatalf("chat results = %+v, %v", chatResults, err)
	}
	message := &Message{Content: []MessageContentBlock{{Type: "tool_use", ID: "msg-1", Name: "fetch", Input: map[string]interface{}{}}}}
	messageResults, _, err := session.HandleMessageToolCalls(ctx, message)
	if err != nil || len(messageResults) != 1 || !strings.Contains(string(messageResults[0].Content), `"tool_use_id":"msg-1"`) {
		t.Fatalf("message results = %+v, %v", messageResults, err)
	}
	callID, name, args := "resp-1", "fetch", `{}`
	modelResponse := &ResponsesResponse{Output: []ResponseOutputItem{{Type: "function_call", CallID: &callID, Name: &name, Arguments: &args}}}
	responseResults, _, err := session.HandleResponseToolCalls(ctx, modelResponse)
	if err != nil || len(responseResults) != 1 || responseResults[0].Type != "function_call_output" || responseResults[0].CallID != callID {
		t.Fatalf("response results = %+v, %v", responseResults, err)
	}
	aliases := []ActionGatewayToolCall{{ID: "alias", Name: actionGatewayInvokeTool, Arguments: json.RawMessage(`{"tools":[{"function":{"name":"fetch","arguments":"{\"url\":\"example.com\"}"}}]}`)}}
	aliasResults, _, err := session.ExecuteToolCalls(ctx, aliases, "")
	if err != nil || len(aliasResults) != 1 || !strings.Contains(string(aliasResults[0]), "fetch") {
		t.Fatalf("alias results = %s, %v", aliasResults, err)
	}
}

func TestActionGatewayAuthenticationAndErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer sdk-token" {
			t.Errorf("missing PAT on %s", request.URL.Path)
		}
		switch request.URL.Path {
		case actionGatewayPath + "/sessions":
			_, _ = io.WriteString(writer, `{"session":{"sessionUrn":"urn:abc"},"mcpUrl":"https://actions.do-ai.run/mcp/session/abc"}`)
		case "/mcp/session/abc":
			writer.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(writer, "data: {\"event\":\"ping\"}\n\ndata: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"tools\":[]}}\n\n")
		default:
			writer.WriteHeader(http.StatusForbidden)
			_, _ = io.WriteString(writer, `{"message":"forbidden"}`)
		}
	}))
	defer server.Close()
	authClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "sdk-token"}))
	client, err := New(authClient, SetBaseURL(server.URL), WithActionGatewayMCPBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	session, _, err := client.ActionGateway.Sessions.CreateRuntime(context.Background(), &ActionGatewaySessionCreateRequest{ActorID: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	tools, _, err := session.Tools.List(context.Background(), false)
	if err != nil || len(tools) != 0 {
		t.Fatalf("SSE list = %+v, %v", tools, err)
	}
	_, response, err := client.ActionGateway.Users.Get(context.Background(), "missing")
	var apiError *ErrorResponse
	if !errors.As(err, &apiError) || response.StatusCode != http.StatusForbidden || apiError.Message != "forbidden" {
		t.Fatalf("HTTP error = %+v, %v", response, err)
	}
}

func TestActionGatewayRuntimeFailures(t *testing.T) {
	client := NewClient(nil)
	if _, _, err := client.ActionGateway.Sessions.CreateRuntime(context.Background(), &ActionGatewaySessionCreateRequest{}); err == nil {
		t.Fatal("expected missing actor error")
	}
	if _, err := New(nil, WithActionGatewayMCPBaseURL("http://example.com")); err == nil {
		t.Fatal("expected HTTPS requirement")
	}
	if _, err := newActionGatewaySession(client, &ActionGatewaySessionCreateResponse{Session: &ActionGatewaySessionRecord{SessionURN: "urn:x"}, MCPURL: "http://actions.do-ai.run/mcp"}, "alice"); err == nil {
		t.Fatal("expected rejection of insecure returned URL")
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/mcp/session/abc" {
			http.Redirect(writer, request, "https://example.com/steal", http.StatusTemporaryRedirect)
			return
		}
		_, _ = io.WriteString(writer, `{"session":{"sessionUrn":"urn:abc"},"mcpUrl":"https://actions.do-ai.run/mcp/session/abc"}`)
	}))
	defer server.Close()
	client, err := New(server.Client(), SetBaseURL(server.URL), WithActionGatewayMCPBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	session, _, err := client.ActionGateway.Sessions.CreateRuntime(context.Background(), &ActionGatewaySessionCreateRequest{ActorID: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = session.Tools.List(context.Background(), false); err == nil || !strings.Contains(err.Error(), "cross-origin") {
		t.Fatalf("expected redirect rejection, got %v", err)
	}
}

func TestActionGatewaySSEIncremental(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var call struct {
			ID int64 `json:"id"`
		}
		if err := json.NewDecoder(request.Body).Decode(&call); err != nil {
			t.Error(err)
			return
		}
		writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		_, _ = fmt.Fprintf(writer, "data: {\"event\":\"ping\"}\n\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"notifications/progress\"}\n\ndata: {\"jsonrpc\":\"2.0\",\"id\":%d,\"result\":{\"tools\":[{\"name\":\"current\"}]}}\n\n", call.ID)
		writer.(http.Flusher).Flush()
		<-request.Context().Done()
	}))
	defer server.Close()
	client, err := New(server.Client(), WithActionGatewayMCPBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	session, err := newActionGatewaySession(client, &ActionGatewaySessionCreateResponse{Session: &ActionGatewaySessionRecord{SessionURN: "urn:abc"}, MCPURL: "https://actions.do-ai.run/mcp/session/abc"}, "alice")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	tools, response, err := session.Tools.List(ctx, true)
	if err != nil || response == nil || len(tools) != 1 || tools[0].Name != "current" {
		t.Fatalf("incremental SSE tools = %+v, response = %+v, err = %v", tools, response, err)
	}
}

func TestActionGatewayRPCResponseIDs(t *testing.T) {
	cases := []struct {
		name, contentType, body, message string
	}{
		{"wrong JSON result", "application/json", `{"jsonrpc":"2.0","id":2,"result":{"tools":[]}}`, "mismatched"},
		{"missing JSON ID", "application/json", `{"jsonrpc":"2.0","result":{"tools":[]}}`, "mismatched"},
		{"string JSON ID", "application/json", `{"jsonrpc":"2.0","id":"1","result":{"tools":[]}}`, "mismatched"},
		{"wrong JSON error", "application/json", `{"jsonrpc":"2.0","id":2,"error":{"code":-32001,"message":"unavailable"}}`, "mismatched"},
		{"wrong SSE result", "text/event-stream", "data: {\"jsonrpc\":\"2.0\",\"id\":2,\"result\":{\"tools\":[]}}\n\n", "mismatched"},
		{"wrong SSE error", "text/event-stream", "data: {\"jsonrpc\":\"2.0\",\"id\":2,\"error\":{\"code\":-32001,\"message\":\"unavailable\"}}\n\n", "mismatched"},
		{"missing SSE ID", "text/event-stream", "data: {\"jsonrpc\":\"2.0\",\"result\":{\"tools\":[]}}\n\n", "invalid JSON-RPC response ID"},
		{"empty SSE", "text/event-stream", "data: {\"event\":\"ping\"}\n\n", "without a matching"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", test.contentType)
				_, _ = io.WriteString(writer, test.body)
			}))
			defer server.Close()
			client, err := New(server.Client(), WithActionGatewayMCPBaseURL(server.URL))
			if err != nil {
				t.Fatal(err)
			}
			session, err := newActionGatewaySession(client, &ActionGatewaySessionCreateResponse{Session: &ActionGatewaySessionRecord{SessionURN: "urn:abc"}, MCPURL: "https://actions.do-ai.run/mcp/session/abc"}, "alice")
			if err != nil {
				t.Fatal(err)
			}
			_, response, err := session.Tools.List(context.Background(), true)
			var protocolError *ActionGatewayProtocolError
			if !errors.As(err, &protocolError) || !strings.Contains(protocolError.Message, test.message) || response == nil || response.StatusCode != http.StatusOK {
				t.Fatalf("protocol error = %v, response = %+v", err, response)
			}
		})
	}
}
