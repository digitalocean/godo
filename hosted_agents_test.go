package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var hostedAgentSession = HostedAgentSession{
	SessionID: "sess-abc123",
	TeamID:    42,
	AgentKind: HostedAgentKindClaudeCode,
	Status:    HostedAgentSessionStatusReady,
	SandboxID: "sandbox-xyz",
	CreatedAt: Timestamp{Time: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)},
	LastEventAt: Timestamp{Time: time.Date(2026, 3, 1, 12, 5, 0, 0, time.UTC)},
	RepoHint:  "digitalocean/godo",
	ProviderAuth: map[string]HostedAgentProviderAuthState{
		"github": HostedAgentProviderAuthStateAuthorized,
	},
}

var hostedAgentSessionJSON = `
{
	"session_id": "sess-abc123",
	"team_id": 42,
	"agent_kind": "AGENT_KIND_CLAUDE_CODE",
	"status": "SESSION_STATUS_READY",
	"sandbox_id": "sandbox-xyz",
	"created_at": "2026-03-01T12:00:00Z",
	"last_event_at": "2026-03-01T12:05:00Z",
	"repo_hint": "digitalocean/godo",
	"provider_auth": {
		"github": "PROVIDER_AUTH_STATE_AUTHORIZED"
	}
}
`

func TestHostedAgents_CreateSession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentSessionCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, HostedAgentKindClaudeCode, body.AgentKind)
		assert.Equal(t, "digitalocean/godo", body.RepoHint)
		fmt.Fprintf(w, `{"session":%s}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.CreateSession(ctx, &HostedAgentSessionCreateRequest{
		AgentKind: HostedAgentKindClaudeCode,
		RepoHint:  "digitalocean/godo",
	})
	require.NoError(t, err)
	assert.Equal(t, hostedAgentSession, *got)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_ListSessions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "SESSION_STATUS_READY", r.URL.Query().Get("status"))
		assert.Equal(t, "25", r.URL.Query().Get("page_size"))
		fmt.Fprintf(w, `{"sessions":[%s],"next_page_token":""}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.ListSessions(ctx, &HostedAgentSessionListOptions{
		Status:   HostedAgentSessionStatusReady,
		PageSize: 25,
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 1)
	assert.Equal(t, hostedAgentSession, got.Sessions[0])
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_GetSession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"session":%s}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.GetSession(ctx, "sess-abc123")
	require.NoError(t, err)
	assert.Equal(t, hostedAgentSession, *got)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_DestroySession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.DestroySession(ctx, "sess-abc123")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHostedAgents_SendInput(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/input", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentSendInputRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "fix the failing test", body.Text)
		fmt.Fprint(w, `{"run_id":"run-001"}`)
	})

	got, resp, err := client.HostedAgents.SendInput(ctx, "sess-abc123", &HostedAgentSendInputRequest{
		Text: "fix the failing test",
	})
	require.NoError(t, err)
	assert.Equal(t, "run-001", got.RunID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_ResolveHITL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/hitl/req-001", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentResolveHITLRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, HostedAgentHITLOutcomeApprove, body.Outcome)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.ResolveHITL(ctx, "sess-abc123", "req-001", &HostedAgentResolveHITLRequest{
		Outcome: HostedAgentHITLOutcomeApprove,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHostedAgents_StartOAuthFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/oauth/github", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"authorize_url":"https://github.com/login/oauth/authorize?state=abc","flow_kind":"OAUTH_FLOW_KIND_WEB_CALLBACK"}`)
	})

	got, resp, err := client.HostedAgents.StartOAuthFlow(ctx, "sess-abc123", "github", nil)
	require.NoError(t, err)
	assert.Equal(t, HostedAgentOAuthFlowKindWebCallback, got.FlowKind)
	assert.Contains(t, got.AuthorizeURL, "github.com")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_StreamSession(t *testing.T) {
	setup()
	defer teardown()

	const eventJSON = `{"event_id":"ev-1","session_id":"sess-abc123","team_id":42,"at":"2026-03-01T12:01:00Z","kind":"EVENT_KIND_TOKEN_CHUNK","payload":{"text":"hello"}}`

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/stream", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))
		assert.Equal(t, "ev-0", r.URL.Query().Get("replay_from"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, ": connected\n\n")
		fmt.Fprintf(w, "id: ev-1\nevent: EVENT_KIND_TOKEN_CHUNK\ndata: %s\n\n", eventJSON)
	})

	stream, resp, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		ReplayFrom: "ev-0",
	})
	require.NoError(t, err)
	require.NotNil(t, stream)
	defer stream.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	require.True(t, stream.Next())
	assert.Equal(t, HostedAgentEventKindTokenChunk, stream.Current().Kind)
	assert.Equal(t, "ev-1", stream.Current().EventID)
	assert.NoError(t, stream.Err())
	assert.False(t, stream.Next())
}

func TestHostedAgents_ListSessions_PageToken(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "cursor-abc", r.URL.Query().Get("page_token"))
		fmt.Fprintf(w, `{"sessions":[],"next_page_token":"cursor-def"}`)
	})

	got, resp, err := client.HostedAgents.ListSessions(ctx, &HostedAgentSessionListOptions{
		PageToken: "cursor-abc",
	})
	require.NoError(t, err)
	assert.Empty(t, got.Sessions)
	assert.Equal(t, "cursor-def", got.NextPageToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_ExecInSandbox(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/sandbox/exec", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentSandboxExecRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, []string{"echo", "hello"}, body.Argv)
		fmt.Fprint(w, `{"exit_code":0,"stdout":"hello\n"}`)
	})

	got, resp, err := client.HostedAgents.ExecInSandbox(ctx, "sess-abc123", &HostedAgentSandboxExecRequest{
		Argv: []string{"echo", "hello"},
	})
	require.NoError(t, err)
	assert.Equal(t, 0, got.ExitCode)
	assert.Equal(t, "hello\n", got.Stdout)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_StreamSession_ReplayOnly(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/stream", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "true", r.URL.Query().Get("replay_only"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ": replay only\n\n")
	})

	stream, resp, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		ReplayOnly: true,
	})
	require.NoError(t, err)
	defer stream.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.False(t, stream.Next())
}

func TestHostedAgents_ValidationErrors(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.CreateSession(ctx, nil)
	require.EqualError(t, err, "hosted agents: create request is required")

	_, _, err = client.HostedAgents.CreateSession(ctx, &HostedAgentSessionCreateRequest{})
	require.EqualError(t, err, "hosted agents: agent_kind is required")

	_, _, err = client.HostedAgents.GetSession(ctx, "")
	require.EqualError(t, err, "hosted agents: session id is required")

	_, err = client.HostedAgents.DestroySession(ctx, "")
	require.EqualError(t, err, "hosted agents: session id is required")

	_, _, err = client.HostedAgents.SendInput(ctx, "sess-abc123", nil)
	require.EqualError(t, err, "hosted agents: input is required")

	_, _, err = client.HostedAgents.SendInput(ctx, "sess-abc123", &HostedAgentSendInputRequest{})
	require.EqualError(t, err, "hosted agents: text is required")

	_, err = client.HostedAgents.ResolveHITL(ctx, "", "req-001", &HostedAgentResolveHITLRequest{
		Outcome: HostedAgentHITLOutcomeApprove,
	})
	require.EqualError(t, err, "hosted agents: session id is required")

	_, err = client.HostedAgents.ResolveHITL(ctx, "sess-abc123", "req-001", &HostedAgentResolveHITLRequest{})
	require.EqualError(t, err, "hosted agents: outcome is required")

	_, _, err = client.HostedAgents.StartOAuthFlow(ctx, "sess-abc123", "", nil)
	require.EqualError(t, err, "hosted agents: provider is required")

	_, _, err = client.HostedAgents.ExecInSandbox(ctx, "sess-abc123", &HostedAgentSandboxExecRequest{})
	require.EqualError(t, err, "hosted agents: argv is required")
}

func TestHostedAgents_GetSession_EmptyBody(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	})

	_, resp, err := client.HostedAgents.GetSession(ctx, "sess-abc123")
	require.EqualError(t, err, "hosted agents: get session returned no session")
	require.NotNil(t, resp)
}
