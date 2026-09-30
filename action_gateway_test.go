package godo

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestActionGatewayPublicOperations(t *testing.T) {
	type operation struct {
		resource, name, method, path, query, body string
		args                                      []interface{}
	}
	ctx := context.Background()
	operations := []operation{
		{"Tools", "List", "GET", "/tools", "page=2&per_page=7&toolkit_id=exa", "", []interface{}{&ActionGatewayToolListOptions{ToolkitID: "exa", Page: 2, PerPage: 7}}},
		{"Tools", "ListToolkits", "GET", "/tools/toolkits", "", "", nil},
		{"Tools", "ListProviders", "GET", "/tools/providers", "", "", nil},
		{"Tools", "Search", "GET", "/tools/search", "provider=a&provider=b&query=foo+bar", "", []interface{}{&ActionGatewayToolSearchOptions{Providers: []string{"a", "b"}, Query: "foo bar"}}},
		{"Tools", "SearchProviders", "GET", "/tools/providers/search", "page_token=next", "", []interface{}{&ActionGatewayProviderSearchOptions{PageToken: "next"}}},
		{"Tools", "ListHealth", "GET", "/tools/health", "provider=exa&window=HEALTH_WINDOW_ONE_HOUR", "", []interface{}{&ActionGatewayHealthOptions{Provider: "exa", Window: "HEALTH_WINDOW_ONE_HOUR"}}},
		{"Tools", "ListProviderHealth", "GET", "/tools/health/providers", "", "", []interface{}{(*ActionGatewayHealthOptions)(nil)}},
		{"Tools", "GetHealth", "GET", "/tools/health/tools/exa", "include_history=true", "", []interface{}{"exa", &ActionGatewayToolHealthOptions{IncludeHistory: true}}},
		{"Toolbelts", "Create", "POST", "/toolbelts", "", `{"name":"belt","tools":["exa"]}`, []interface{}{&ActionGatewayToolbeltCreateRequest{Name: "belt", Tools: []string{"exa"}}}},
		{"Toolbelts", "List", "GET", "/toolbelts", "status=active", "", []interface{}{&ActionGatewayToolbeltsListOptions{Status: "active"}}},
		{"Toolbelts", "Search", "GET", "/toolbelts/search", "query=find", "", []interface{}{&ActionGatewayToolbeltsSearchOptions{Query: "find"}}},
		{"Toolbelts", "Get", "GET", "/toolbelts/belt", "version=2", "", []interface{}{"belt", &ActionGatewayGetToolbeltOptions{Version: "2"}}},
		{"Toolbelts", "Delete", "DELETE", "/toolbelts/belt", "", "", []interface{}{"belt"}},
		{"Toolbelts", "AddTools", "POST", "/toolbelts/belt/tools/add", "", `{"tools":["exa"]}`, []interface{}{"belt", &ActionGatewayToolbeltToolsRequest{Tools: []string{"exa"}}}},
		{"Toolbelts", "RemoveTools", "POST", "/toolbelts/belt/tools/remove", "", `{"tools":["exa"]}`, []interface{}{"belt", &ActionGatewayToolbeltToolsRequest{Tools: []string{"exa"}}}},
		{"Toolbelts", "ListProviders", "GET", "/toolbelts/belt/providers", "page=2", "", []interface{}{"belt", &ActionGatewayToolbeltProviderOptions{Page: 2}}},
		{"Toolbelts", "ListProviderTools", "GET", "/toolbelts/belt/providers/exa/tools", "search=web", "", []interface{}{"belt", "exa", &ActionGatewayToolbeltProviderOptions{Search: "web"}}},
		{"OutputViews", "Create", "POST", "/output-views", "", `{"tool":"exa","name":"titles","fields":["title"]}`, []interface{}{&ActionGatewayOutputViewCreateRequest{Tool: "exa", Name: "titles", Fields: []string{"title"}}}},
		{"OutputViews", "Preview", "POST", "/output-views/preview", "", `{"tool":"exa","fields":["title"]}`, []interface{}{&ActionGatewayOutputViewPreviewRequest{Tool: "exa", Fields: []string{"title"}}}},
		{"OutputViews", "List", "GET", "/output-views", "page_token=cursor&tool=exa", "", []interface{}{&ActionGatewayOutputViewListOptions{Tool: "exa", PageToken: "cursor"}}},
		{"OutputViews", "Get", "GET", "/output-views/one", "", "", []interface{}{"one"}},
		{"OutputViews", "Delete", "DELETE", "/output-views/one", "", "", []interface{}{"one"}},
		{"MCPServers", "Create", "POST", "/mcp-servers", "", `{"serverRef":"docs","endpoint":"https://example.com/mcp"}`, []interface{}{&ActionGatewayMCPServerCreateRequest{ServerRef: "docs", Endpoint: "https://example.com/mcp"}}},
		{"MCPServers", "List", "GET", "/mcp-servers", "", "", nil},
		{"MCPServers", "Get", "GET", "/mcp-servers/docs", "", "", []interface{}{"docs"}},
		{"MCPServers", "Update", "PATCH", "/mcp-servers/docs", "", `{"description":"new"}`, []interface{}{"docs", &ActionGatewayMCPServerUpdateRequest{Description: "new"}}},
		{"MCPServers", "Delete", "DELETE", "/mcp-servers/docs", "", "", []interface{}{"docs"}},
		{"MCPServers", "Resync", "POST", "/mcp-servers/docs/resync", "", `{"user_id":"alice"}`, []interface{}{"docs", &ActionGatewayMCPServerResyncRequest{UserID: "alice"}}},
		{"MCPServers", "ListTools", "GET", "/mcp-servers/docs/tools", "", "", []interface{}{"docs"}},
		{"MCPServers", "UpdateTools", "PUT", "/mcp-servers/docs/tools", "", `{"enabledToolSlugs":["one"]}`, []interface{}{"docs", &ActionGatewayMCPServerToolsUpdateRequest{EnabledToolSlugs: []string{"one"}}}},
		{"Connections", "Create", "POST", "/connections", "", `{"provider":"jira","user_id":"alice","credential":{"team_credential":{"credential_id":"cred"}},"network":{"vpc":{"vpc_uuid":"vpc","destinations":[{"host":"internal","port":443,"allowed_ip_cidrs":["10.0.0.0/8"]}]}}}`, []interface{}{&ActionGatewayConnectionCreateRequest{Provider: "jira", UserID: "alice", Credential: &ActionGatewayConnectionCredential{TeamCredential: &ActionGatewayTeamCredential{CredentialID: "cred"}}, Network: &ActionGatewayConnectionCreateNetwork{VPC: &ActionGatewayConnectionCreateVPC{VPCUUID: "vpc", Destinations: []ActionGatewayConnectionDestination{{Host: "internal", Port: 443, AllowedIPCIDRs: []string{"10.0.0.0/8"}}}}}}}},
		{"Connections", "List", "GET", "/connections", "provider=jira&sort_direction=desc", "", []interface{}{&ActionGatewayConnectionListOptions{Provider: "jira", SortDirection: "desc"}}},
		{"Connections", "Get", "GET", "/connections/id", "", "", []interface{}{"id"}},
		{"Connections", "Delete", "DELETE", "/connections/id", "", "", []interface{}{"id"}},
		{"Sessions", "Create", "POST", "/sessions", "", `{"name":"test","tools":[]}`, []interface{}{&ActionGatewaySessionCreateRequest{Name: "test", Tools: []string{}}}},
		{"Sessions", "List", "GET", "/sessions", "end_user_id=alice&page=2", "", []interface{}{&ActionGatewaySessionListOptions{EndUserID: "alice", Page: 2}}},
		{"Sessions", "Search", "GET", "/sessions/search", "page_token=cursor&query=alice", "", []interface{}{&ActionGatewaySessionSearchOptions{Query: "alice", PageToken: "cursor"}}},
		{"Sessions", "Delete", "DELETE", "/sessions/do:managed_agent_session:abc", "", "", []interface{}{"do:managed_agent_session:abc"}},
		{"Users", "List", "GET", "/users", "sort=created_at", "", []interface{}{&ActionGatewayUsersListOptions{Sort: "created_at"}}},
		{"Users", "Get", "GET", "/users/alice%2Fbob", "", "", []interface{}{"alice/bob"}},
		{"ActorLimits", "Get", "GET", "/actors/alice/limits", "", "", []interface{}{"alice"}},
		{"ActorLimits", "Set", "POST", "/actors/alice/limits:set", "", `{"overrides":[{"category":"LIMIT_CATEGORY_WEB_SEARCH_REQUESTS_PER_MINUTE","requests_per_minute":"10"}]}`, []interface{}{"alice", &ActionGatewaySetLimitsRequest{Overrides: []ActionGatewayLimitOverride{{Category: "LIMIT_CATEGORY_WEB_SEARCH_REQUESTS_PER_MINUTE", RequestsPerMinute: "10"}}}}},
		{"ActorLimits", "Clear", "POST", "/actors/alice/limits:clear", "", `{"categories":["LIMIT_CATEGORY_WEB_SEARCH_REQUESTS_PER_MINUTE"]}`, []interface{}{"alice", &ActionGatewayClearLimitsRequest{Categories: []string{"LIMIT_CATEGORY_WEB_SEARCH_REQUESTS_PER_MINUTE"}}}},
	}
	if len(operations) != 43 {
		t.Fatalf("expected all 43 public operations, got %d", len(operations))
	}
	for _, operation := range operations {
		t.Run(operation.resource+"/"+operation.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != operation.method || request.URL.EscapedPath() != actionGatewayPath+operation.path || request.URL.RawQuery != operation.query {
					t.Errorf("request = %s %s?%s, want %s %s?%s", request.Method, request.URL.EscapedPath(), request.URL.RawQuery, operation.method, actionGatewayPath+operation.path, operation.query)
				}
				if operation.body != "" {
					payload, _ := io.ReadAll(request.Body)
					var got, want interface{}
					if json.Unmarshal(payload, &got) != nil || json.Unmarshal([]byte(operation.body), &want) != nil || !reflect.DeepEqual(got, want) {
						t.Errorf("body = %s, want %s", payload, operation.body)
					}
				}
				writer.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(writer, "{}")
			}))
			defer server.Close()
			client := NewClient(server.Client())
			client.BaseURL, _ = url.Parse(server.URL)
			service := reflect.ValueOf(client.ActionGateway).Elem().FieldByName(operation.resource)
			method := service.MethodByName(operation.name)
			if !method.IsValid() {
				t.Fatalf("missing %s.%s", operation.resource, operation.name)
			}
			args := []reflect.Value{reflect.ValueOf(ctx)}
			for _, arg := range operation.args {
				args = append(args, reflect.ValueOf(arg))
			}
			results := method.Call(args)
			if errValue := results[len(results)-1]; !errValue.IsNil() {
				t.Fatal(errValue.Interface())
			}
		})
	}
}

func TestActionGatewayEmptyPathArguments(t *testing.T) {
	client := NewClient(nil)
	ctx := context.Background()
	cases := []struct {
		name, argument string
		call           func() error
	}{
		{"tool slug", "slug", func() error { _, _, err := client.ActionGateway.Tools.GetHealth(ctx, "", nil); return err }},
		{"toolbelt name", "name", func() error { _, _, err := client.ActionGateway.Toolbelts.Get(ctx, "", nil); return err }},
		{"toolbelt provider", "provider", func() error {
			_, _, err := client.ActionGateway.Toolbelts.ListProviderTools(ctx, "belt", "", nil)
			return err
		}},
		{"view ID", "viewID", func() error { _, _, err := client.ActionGateway.OutputViews.Get(ctx, ""); return err }},
		{"server ref", "ref", func() error { _, _, err := client.ActionGateway.MCPServers.Get(ctx, ""); return err }},
		{"connection ID", "id", func() error { _, _, err := client.ActionGateway.Connections.Get(ctx, ""); return err }},
		{"session URN", "urn", func() error { _, err := client.ActionGateway.Sessions.Delete(ctx, ""); return err }},
		{"user ID", "userID", func() error { _, _, err := client.ActionGateway.Users.Get(ctx, ""); return err }},
		{"actor ID", "actorID", func() error { _, _, err := client.ActionGateway.ActorLimits.Get(ctx, ""); return err }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var argumentError *ArgError
			if err := test.call(); !errors.As(err, &argumentError) || argumentError.arg != test.argument {
				t.Fatalf("error = %v, want argument %q", err, test.argument)
			}
		})
	}
}

func TestActionGatewayWireShapes(t *testing.T) {
	var session ActionGatewaySessionRecord
	if err := json.Unmarshal([]byte(`{"sessionUrn":"do:managed_agent_session:x","actorId":"alice","tools":{"references":[]},"owning_user_numeric_id":"18446744073709551615"}`), &session); err != nil {
		t.Fatal(err)
	}
	if session.ActorID != "alice" || session.OwningUserNumericID != "18446744073709551615" || session.Tools == nil || session.Tools.References == nil {
		t.Fatalf("unexpected session: %+v", session)
	}
	var server ActionGatewayMCPServer
	if err := json.Unmarshal([]byte(`{"serverRef":"docs","toolCount":3,"lastSyncedAt":"","oauth_authorization_ttl_seconds":"86400"}`), &server); err != nil {
		t.Fatal(err)
	}
	if server.ServerRef != "docs" || server.OAuthAuthorizationTTLSeconds != "86400" || server.ToolCount != 3 {
		t.Fatalf("unexpected MCP server: %+v", server)
	}
	for _, tools := range [][]string{nil, {}} {
		body, err := json.Marshal(ActionGatewaySessionCreateRequest{Name: "x", Tools: tools})
		if err != nil {
			t.Fatal(err)
		}
		if (tools == nil) == strings.Contains(string(body), `"tools"`) {
			t.Fatalf("tools selection not preserved: %s", body)
		}
	}
}

func TestActionGatewayResponseDecoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case actionGatewayPath + "/tools":
			_, _ = io.WriteString(writer, `{"version":"v1","tools":[{"toolSlug":"exa_web_search","inputSchema":{"type":"object"}}],"definitions":[{"toolId":"t1","reliability":{"maxOutputBytes":"2097152","retry":{"backoffPolicy":{"initialMs":"200","multiplier":2}}}}],"pagination":{"page":2,"per_page":20,"total":31}}`)
		case actionGatewayPath + "/toolbelts/belt":
			_, _ = io.WriteString(writer, `{"toolbelt":{"name":"belt","version":"2"},"tool_details":[{"tool_slug":"exa","version":1}],"next_page_token":"next"}`)
		case actionGatewayPath + "/mcp-servers/docs":
			_, _ = io.WriteString(writer, `{"mcpServer":{"serverRef":"docs","lastSyncedAt":"","createdAt":"2026-09-20T15:00:00Z","oauth_authorization_ttl_seconds":"86400"}}`)
		case actionGatewayPath + "/mcp-servers/docs/resync":
			writer.WriteHeader(http.StatusAccepted)
			_, _ = io.WriteString(writer, `{"pending":true,"authorization":{"status":"requires_authorization","verification_code":"code"}}`)
		case actionGatewayPath + "/connections/id":
			_, _ = io.WriteString(writer, `{"connection":{"id":"id","owning_user_numeric_id":"18446744073709551615"},"authorization":{"connect_url":"https://cloud.digitalocean.com/connect","expires_at":"2026-09-20T15:00:00Z"}}`)
		case actionGatewayPath + "/output-views/one":
			_, _ = io.WriteString(writer, `{"view":{"view_id":"one","output_schema":{"type":"object","properties":{"answer":{"type":"string"}}},"audit":{"createdAt":"2026-09-20T15:00:00Z"}}}`)
		case actionGatewayPath + "/sessions":
			_, _ = io.WriteString(writer, `{"sessions":[{"sessionUrn":"urn:abc","policy":{"defaultAction":"ask"},"tools":{"references":[]}}],"pagination":{"page":1,"per_page":20,"total":1}}`)
		case actionGatewayPath + "/users/alice":
			_, _ = io.WriteString(writer, `{"user":{"user_id":"alice","sessions":[{"session_urn":"urn:abc"}],"connections":[{"credential_kind":"team_api_key"}]}}`)
		case actionGatewayPath + "/actors/alice/limits":
			_, _ = io.WriteString(writer, `{"configured_limits":[{"category":"LIMIT_CATEGORY_WEB_SEARCH_REQUESTS_PER_MINUTE","requests_per_minute":"10"}],"effective_limits":{"native_tool_calls_requests_per_minute":"600"}}`)
		default:
			t.Errorf("unexpected route %s", request.URL.Path)
		}
	}))
	defer server.Close()
	client := NewClient(server.Client())
	client.BaseURL, _ = url.Parse(server.URL)
	ctx := context.Background()
	tools, _, err := client.ActionGateway.Tools.List(ctx, nil)
	if err != nil || tools.Pagination.Total != 31 || tools.Tools[0].ToolSlug != "exa_web_search" || tools.Definitions[0].Reliability.MaxOutputBytes != "2097152" {
		t.Fatalf("tools = %+v, %v", tools, err)
	}
	belt, _, err := client.ActionGateway.Toolbelts.Get(ctx, "belt", nil)
	if err != nil || belt.NextPageToken != "next" || belt.ToolDetails[0].Version != 1 {
		t.Fatalf("toolbelt = %+v, %v", belt, err)
	}
	mcp, _, err := client.ActionGateway.MCPServers.Get(ctx, "docs")
	if err != nil || mcp.MCPServer.LastSyncedAt != "" || mcp.MCPServer.OAuthAuthorizationTTLSeconds != "86400" {
		t.Fatalf("MCP server = %+v, %v", mcp, err)
	}
	resync, response, err := client.ActionGateway.MCPServers.Resync(ctx, "docs", &ActionGatewayMCPServerResyncRequest{})
	if err != nil || response.StatusCode != http.StatusAccepted || !resync.Pending || resync.Authorization.VerificationCode != "code" {
		t.Fatalf("resync = %+v, %v", resync, err)
	}
	connection, _, err := client.ActionGateway.Connections.Get(ctx, "id")
	if err != nil || connection.Connection.OwningUserNumericID != "18446744073709551615" || connection.Authorization.ExpiresAt == nil {
		t.Fatalf("connection = %+v, %v", connection, err)
	}
	view, _, err := client.ActionGateway.OutputViews.Get(ctx, "one")
	if err != nil || view.View.Audit.CreatedAt == nil || !strings.Contains(string(view.View.OutputSchema), "answer") {
		t.Fatalf("view = %+v, %v", view, err)
	}
	sessions, _, err := client.ActionGateway.Sessions.List(ctx, nil)
	if err != nil || sessions.Pagination.Total != 1 || sessions.Sessions[0].Policy.DefaultAction != "ask" {
		t.Fatalf("sessions = %+v, %v", sessions, err)
	}
	user, _, err := client.ActionGateway.Users.Get(ctx, "alice")
	if err != nil || user.User.Sessions[0].SessionURN != "urn:abc" {
		t.Fatalf("user = %+v, %v", user, err)
	}
	limits, _, err := client.ActionGateway.ActorLimits.Get(ctx, "alice")
	if err != nil || limits.ConfiguredLimits[0].RequestsPerMinute != "10" || limits.EffectiveLimits.NativeToolCallsRequestsPerMinute != "600" {
		t.Fatalf("limits = %+v, %v", limits, err)
	}
}
