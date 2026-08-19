package godo

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

var hostedAgentSession = HostedAgentSession{
	SessionID:   "sess-abc123",
	Name:        "godo-fixture",
	AgentKind:   HostedAgentKindClaudeCode,
	Status:      HostedAgentSessionStatusReady,
	CreatedAt:   Timestamp{Time: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)},
	LastEventAt: Timestamp{Time: time.Date(2026, 3, 1, 12, 5, 0, 0, time.UTC)},
	RepoHint:    "digitalocean/godo",
	ProviderAuth: map[string]HostedAgentProviderAuthState{
		"github": HostedAgentProviderAuthStateAuthorized,
	},
}

var hostedAgentSessionJSON = `
{
	"session_id": "sess-abc123",
	"name": "godo-fixture",
	"agent_kind": "AGENT_KIND_CLAUDE_CODE",
	"status": "SESSION_STATUS_READY",
	"created_at": "2026-03-01T12:00:00Z",
	"last_event_at": "2026-03-01T12:05:00Z",
	"repo_hint": "digitalocean/godo",
	"provider_auth": {
		"github": "PROVIDER_AUTH_STATE_AUTHORIZED"
	}
}
`

var hostedAgentOpenAISession = HostedAgentSession{
	SessionID:           "sess-openai-1",
	Name:                "openai-codex-session",
	AgentKind:           HostedAgentKindOpenAICodex,
	Status:              HostedAgentSessionStatusReady,
	CreatedAt:           Timestamp{Time: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)},
	LastEventAt:         Timestamp{Time: time.Date(2026, 7, 22, 12, 2, 0, 0, time.UTC)},
	OpenAISessionID:     "sess_a91f3",
	OpenAIEnvironmentID: "env_abc123",
}

var hostedAgentOpenAISessionJSON = `
{
	"session_id": "sess-openai-1",
	"name": "openai-codex-session",
	"agent_kind": "AGENT_KIND_OPENAI_CODEX",
	"status": "SESSION_STATUS_READY",
	"created_at": "2026-07-22T12:00:00Z",
	"last_event_at": "2026-07-22T12:02:00Z",
	"openai_session_id": "sess_a91f3",
	"openai_environment_id": "env_abc123"
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

func TestHostedAgents_CreateSession_WithOrigin(t *testing.T) {
	setup()
	defer teardown()

	wantCreate := &HostedAgentSessionCreateRequest{
		AgentKind: HostedAgentKindClaudeCode,
		RepoHint:  "github.com/digitalocean/example",
		Origin: &HostedAgentSessionOriginRequest{
			Product:    HostedAgentSessionOriginProductSimulation,
			ResourceID: "sim-run-123",
		},
	}

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		var got HostedAgentSessionCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, wantCreate.AgentKind, got.AgentKind)
		assert.Equal(t, wantCreate.RepoHint, got.RepoHint)
		require.NotNil(t, got.Origin)
		assert.Equal(t, HostedAgentSessionOriginProductSimulation, got.Origin.Product)
		assert.Equal(t, "sim-run-123", got.Origin.ResourceID)

		// Request wire must not include verified.
		raw, err := json.Marshal(got.Origin)
		require.NoError(t, err)
		assert.NotContains(t, string(raw), "verified")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"session": {
				"session_id": "sess-1",
				"name": "session-claude-code-2026-07-27",
				"agent_kind": "AGENT_KIND_CLAUDE_CODE",
				"status": "SESSION_STATUS_PROVISIONING",
				"created_at": "2026-07-27T12:00:00Z",
				"last_event_at": "2026-07-27T12:00:00Z",
				"repo_hint": "github.com/digitalocean/example",
				"origin": {
					"product": "simulation",
					"resource_id": "sim-run-123",
					"verified": true
				}
			}
		}`)
	})

	session, _, err := client.HostedAgents.CreateSession(ctx, wantCreate)
	require.NoError(t, err)
	require.NotNil(t, session.Origin)
	assert.Equal(t, HostedAgentSessionOriginProductSimulation, session.Origin.Product)
	assert.Equal(t, "sim-run-123", session.Origin.ResourceID)
	assert.True(t, session.Origin.Verified)
}

func TestHostedAgents_CreateSession_DirectOmitsOrigin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		_, hasOrigin := raw["origin"]
		assert.False(t, hasOrigin, "omitted Origin should not appear in JSON body")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"session": {
				"session_id": "sess-direct",
				"agent_kind": "AGENT_KIND_OPENCODE",
				"status": "SESSION_STATUS_READY",
				"created_at": "2026-07-27T12:00:00Z",
				"last_event_at": "2026-07-27T12:00:00Z",
				"origin": {
					"product": "direct",
					"verified": true
				}
			}
		}`)
	})

	session, _, err := client.HostedAgents.CreateSession(ctx, &HostedAgentSessionCreateRequest{
		AgentKind: HostedAgentKindOpenCode,
	})
	require.NoError(t, err)
	require.NotNil(t, session.Origin)
	assert.Equal(t, HostedAgentSessionOriginProductDirect, session.Origin.Product)
	assert.Empty(t, session.Origin.ResourceID)
	assert.True(t, session.Origin.Verified)
}

func TestHostedAgents_GetSession_WithOrigin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-eval", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"session": {
				"session_id": "sess-eval",
				"agent_kind": "AGENT_KIND_CODEX_CLI",
				"status": "SESSION_STATUS_READY",
				"created_at": "2026-07-27T12:00:00Z",
				"last_event_at": "2026-07-27T12:05:00Z",
				"origin": {
					"product": "evaluation",
					"resource_id": "eval-456",
					"verified": true
				}
			}
		}`)
	})

	session, _, err := client.HostedAgents.GetSession(ctx, "sess-eval")
	require.NoError(t, err)
	require.NotNil(t, session.Origin)
	assert.Equal(t, HostedAgentSessionOriginProductEvaluation, session.Origin.Product)
	assert.Equal(t, "eval-456", session.Origin.ResourceID)
	assert.True(t, session.Origin.Verified)
}

func TestHostedAgentSessionOriginRequest_JSONOmitsVerified(t *testing.T) {
	body, err := json.Marshal(&HostedAgentSessionOriginRequest{
		Product:    HostedAgentSessionOriginProductEvaluation,
		ResourceID: "eval-1",
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"product":"evaluation","resource_id":"eval-1"}`, string(body))
}

func TestHostedAgentSessionOrigin_JSONIncludesVerified(t *testing.T) {
	body, err := json.Marshal(&HostedAgentSessionOrigin{
		Product:  HostedAgentSessionOriginProductDirect,
		Verified: true,
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"product":"direct","verified":true}`, string(body))
}

// The sessions API carries no team/tenant field, so a round-tripped session must
// not invent one.
func TestHostedAgentSession_JSONOmitsTeamID(t *testing.T) {
	body, err := json.Marshal(&hostedAgentSession)
	require.NoError(t, err)

	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &fields))
	assert.NotContains(t, fields, "team_id")
}

// A server still sending team_id must not break the decode.
func TestHostedAgentSession_IgnoresServerTeamID(t *testing.T) {
	const sessionJSON = `{
		"session_id": "sess-legacy",
		"team_id": 42,
		"agent_kind": "AGENT_KIND_CODEX_CLI",
		"status": "SESSION_STATUS_READY",
		"created_at": "2026-08-13T21:01:32Z",
		"last_event_at": "2026-08-13T21:04:50Z"
	}`

	var session HostedAgentSession
	require.NoError(t, json.Unmarshal([]byte(sessionJSON), &session))
	assert.Equal(t, "sess-legacy", session.SessionID)
	assert.Equal(t, HostedAgentSessionStatusReady, session.Status)
}

func TestHostedAgents_CreateSessionFromManifest(t *testing.T) {
	setup()
	defer teardown()

	const manifest = `apiVersion: agents.digitalocean.com/v1alpha1
kind: Agent
metadata:
  name: opencode-coding
spec:
  runtime:
    adapter: opencode
  sandbox:
    template: coding
`

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "application/x-yaml", r.Header.Get("Content-Type"))
		assert.Empty(t, r.URL.Query().Get("openai_session_id"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "agents.digitalocean.com/v1alpha1")
		assert.Contains(t, string(body), "adapter: opencode")
		fmt.Fprintf(w, `{"session":%s}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.CreateSessionFromManifest(ctx, []byte(manifest), nil)
	require.NoError(t, err)
	assert.Equal(t, hostedAgentSession, *got)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_CreateSessionFromManifest_Empty(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.CreateSessionFromManifest(ctx, []byte("  \n"), nil)
	require.EqualError(t, err, "hosted agents: manifest is required")
}

func TestHostedAgents_CreateSessionFromConfig(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		// The config-backed path is selected by a JSON body carrying only
		// name + config_id.
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		var body HostedAgentSessionFromConfigRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "session-from-config", body.Name)
		assert.Equal(t, "019fb39c-14d9-7080-933e-b9b90e25acda", body.ConfigID)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"session": {
				"session_id": "sess-cfg",
				"name": "session-from-config",
				"agent_kind": "AGENT_KIND_OPENCODE",
				"status": "SESSION_STATUS_PROVISIONING",
				"created_at": "2026-08-01T12:00:00Z",
				"last_event_at": "2026-08-01T12:00:00Z",
				"config_id": "019fb39c-14d9-7080-933e-b9b90e25acda"
			}
		}`)
	})

	session, resp, err := client.HostedAgents.CreateSessionFromConfig(ctx, &HostedAgentSessionFromConfigRequest{
		Name:     "session-from-config",
		ConfigID: "019fb39c-14d9-7080-933e-b9b90e25acda",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, session)
	assert.Equal(t, "sess-cfg", session.SessionID)
	assert.Equal(t, "session-from-config", session.Name)
	assert.Equal(t, "019fb39c-14d9-7080-933e-b9b90e25acda", session.ConfigID)
}

func TestHostedAgents_CreateSessionFromConfig_Validation(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.CreateSessionFromConfig(ctx, nil)
	require.Error(t, err)

	_, _, err = client.HostedAgents.CreateSessionFromConfig(ctx, &HostedAgentSessionFromConfigRequest{ConfigID: "019fb39c-14d9-7080-933e-b9b90e25acda"})
	require.Error(t, err)

	_, _, err = client.HostedAgents.CreateSessionFromConfig(ctx, &HostedAgentSessionFromConfigRequest{Name: "session-from-config"})
	require.Error(t, err)
}

func TestHostedAgents_CreateSessionFromManifest_OpenAISessionID(t *testing.T) {
	setup()
	defer teardown()

	const manifest = `apiVersion: agents.digitalocean.com/v1alpha1
kind: Agent
metadata:
  name: openai-codex-session
spec:
  runtime:
    adapter: codex-agentapi
  sandbox:
    template: codex-agentapi
  env:
    CODEX_ENVIRONMENT_ID: env_abc123
    CODEX_API_KEY: sk-test
  secrets:
    - name: CODEX_API_KEY
      source: tenantSecret
  openai:
    agent:
      model: gpt-5.6-sol
      instructions: "Answer the user clearly and concisely."
    environment:
      type: self_hosted
      workspace_directory: /workspace
`

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "application/x-yaml", r.Header.Get("Content-Type"))
		assert.Equal(t, "sess_a91f3", r.URL.Query().Get("openai_session_id"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "adapter: codex-agentapi")
		assert.Contains(t, string(body), "CODEX_ENVIRONMENT_ID: env_abc123")
		assert.NotContains(t, string(body), "${")
		fmt.Fprintf(w, `{"session":%s}`, hostedAgentOpenAISessionJSON)
	})

	got, resp, err := client.HostedAgents.CreateSessionFromManifest(ctx, []byte(manifest), &HostedAgentManifestCreateOptions{
		OpenAISessionID: "sess_a91f3",
	})
	require.NoError(t, err)
	assert.Equal(t, hostedAgentOpenAISession, *got)
	assert.Equal(t, HostedAgentKindOpenAICodex, got.AgentKind)
	assert.Equal(t, "sess_a91f3", got.OpenAISessionID)
	assert.Equal(t, "env_abc123", got.OpenAIEnvironmentID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_GetSession_OpenAIFields(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-openai-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"session":%s}`, hostedAgentOpenAISessionJSON)
	})

	got, resp, err := client.HostedAgents.GetSession(ctx, "sess-openai-1")
	require.NoError(t, err)
	assert.Equal(t, hostedAgentOpenAISession, *got)
	assert.Equal(t, "sess_a91f3", got.OpenAISessionID)
	assert.Equal(t, "env_abc123", got.OpenAIEnvironmentID)
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

func TestHostedAgents_PauseSession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/pause", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, "{}", string(body))
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.PauseSession(ctx, "sess-abc123")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHostedAgents_ResumeSession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/resume", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, "{}", string(body))
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.ResumeSession(ctx, "sess-abc123")
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

// TestHostedAgents_SendInput_CarriesNativeFrame pins that a caller speaking the
// session's own protocol can send the frame its text came from, and that Text
// still rides alongside it for every consumer that reads only the canonical
// field.
func TestHostedAgents_SendInput_CarriesNativeFrame(t *testing.T) {
	setup()
	defer teardown()

	const frame = `{"jsonrpc":"2.0","method":"turn/start","params":{"model":"gpt-5","input":[{"type":"text","text":"fix the failing test"}]}}`

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/input", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		// Bytes are base64 on the wire, per the proto JSON mapping.
		assert.Contains(t, string(raw), base64.StdEncoding.EncodeToString([]byte(frame)))

		var body HostedAgentSendInputRequest
		require.NoError(t, json.Unmarshal(raw, &body))
		assert.Equal(t, "fix the failing test", body.Text)
		assert.Equal(t, frame, string(body.SourceRaw))
		fmt.Fprint(w, `{"run_id":"run-001"}`)
	})

	got, _, err := client.HostedAgents.SendInput(ctx, "sess-abc123", &HostedAgentSendInputRequest{
		Text:      "fix the failing test",
		SourceRaw: []byte(frame),
	})
	require.NoError(t, err)
	assert.Equal(t, "run-001", got.RunID)
}

// TestHostedAgents_SendInput_OmitsNativeFrameWhenUnset keeps the field off the
// wire for every caller that does not speak the agent's protocol.
func TestHostedAgents_SendInput_OmitsNativeFrameWhenUnset(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/input", func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.NotContains(t, string(raw), "source_raw")
		fmt.Fprint(w, `{"run_id":"run-001"}`)
	})

	_, _, err := client.HostedAgents.SendInput(ctx, "sess-abc123", &HostedAgentSendInputRequest{
		Text: "hello",
	})
	require.NoError(t, err)
}

// TestHostedAgents_ResolveHITL_CarriesNativeReply pins the resolve-side raw
// field: the client answers in the agent's own protocol, and Outcome still
// rides along as the coarse verdict the audit trail records.
func TestHostedAgents_ResolveHITL_CarriesNativeReply(t *testing.T) {
	setup()
	defer teardown()

	const reply = `{"id":7,"result":{"action":"accept","content":{"answer":"yes"}}}`

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/hitl/req-001", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentResolveHITLRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, HostedAgentHITLOutcomeApprove, body.Outcome,
			"outcome stays required alongside the native reply")
		assert.Equal(t, reply, string(body.SourceRaw))
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.ResolveHITL(ctx, "sess-abc123", "req-001", &HostedAgentResolveHITLRequest{
		Outcome:   HostedAgentHITLOutcomeApprove,
		SourceRaw: []byte(reply),
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHostedAgents_RelayRequest(t *testing.T) {
	setup()
	defer teardown()

	const (
		request = `{"jsonrpc":"2.0","id":3,"method":"turn/interrupt","params":{"threadId":"th-1"}}`
		reply   = `{"jsonrpc":"2.0","id":3,"result":{}}`
	)

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/request", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentRelayRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, request, string(body.SourceRaw))
		fmt.Fprintf(w, `{"source_raw":%q}`, base64.StdEncoding.EncodeToString([]byte(reply)))
	})

	got, resp, err := client.HostedAgents.RelayRequest(ctx, "sess-abc123", &HostedAgentRelayRequest{
		SourceRaw: []byte(request),
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, reply, string(got.SourceRaw),
		"the agent's reply comes back verbatim")
}

// TestHostedAgents_RelayRequest_ProtocolErrorIsANormalReply: a JSON-RPC error
// object is the agent answering, not an HTTP failure, so it arrives in the
// reply rather than as an error.
func TestHostedAgents_RelayRequest_ProtocolErrorIsANormalReply(t *testing.T) {
	setup()
	defer teardown()

	const rpcErr = `{"jsonrpc":"2.0","id":4,"error":{"code":-32601,"message":"method not found"}}`

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/request", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"source_raw":%q}`, base64.StdEncoding.EncodeToString([]byte(rpcErr)))
	})

	got, _, err := client.HostedAgents.RelayRequest(ctx, "sess-abc123", &HostedAgentRelayRequest{
		SourceRaw: []byte(`{"jsonrpc":"2.0","id":4,"method":"nope"}`),
	})
	require.NoError(t, err)
	assert.Equal(t, rpcErr, string(got.SourceRaw))
}

// TestHostedAgents_RelayRequest_DeclinedMethodYieldsEmptyReply: the adapter
// refusing to forward a method is reported as an empty reply, which callers
// must distinguish from an answer.
func TestHostedAgents_RelayRequest_DeclinedMethodYieldsEmptyReply(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/request", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	})

	got, _, err := client.HostedAgents.RelayRequest(ctx, "sess-abc123", &HostedAgentRelayRequest{
		SourceRaw: []byte(`{"jsonrpc":"2.0","id":5,"method":"thread/close"}`),
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Empty(t, got.SourceRaw)
}

func TestHostedAgents_StreamSession(t *testing.T) {
	setup()
	defer teardown()

	// The server serializes the SPI canonical event envelope: the discriminator
	// is `type` (dot-separated), the body is `data`, and the timestamp is
	// `timestamp`. The envelope's `tenant_id` is not surfaced to callers.
	const eventJSON = `{"event_id":"ev-1","run_id":"run-1","tenant_id":"42","session_id":"sess-abc123","timestamp":"2026-03-01T12:01:00Z","seq":1,"type":"run.token_delta","data":{"text":"hello"}}`

	// The live stream is served by the data plane at .../events, and the resume
	// cursor rides in the standard Last-Event-ID header rather than a query
	// parameter.
	mux.HandleFunc("/v2/agents/sessions/sess-abc123/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))
		assert.Equal(t, "ev-0", r.Header.Get("Last-Event-ID"))
		assert.Empty(t, r.URL.RawQuery)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, ": keepalive\n\n")
		fmt.Fprintf(w, "id: ev-1\nevent: run.token_delta\ndata: %s\n\n", eventJSON)
	})

	stream, resp, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		ReplayFrom: "ev-0",
	})
	require.NoError(t, err)
	require.NotNil(t, stream)
	defer stream.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	require.True(t, stream.Next())
	ev := stream.Current()
	assert.Equal(t, HostedAgentEventKindTokenChunk, ev.Kind)
	assert.Equal(t, "ev-1", ev.EventID)
	assert.Equal(t, "run-1", ev.RunID)
	assert.JSONEq(t, `{"text":"hello"}`, string(ev.Payload))
	assert.NoError(t, stream.Err())
	assert.False(t, stream.Next())
}

// TestHostedAgentEvent_UnmarshalSPIWire pins the SPI canonical envelope decode:
// type->Kind (dot-separated), data->Payload, timestamp->At.
func TestHostedAgentEvent_UnmarshalSPIWire(t *testing.T) {
	const frame = `{"event_id":"ev-9","run_id":"run-7","tenant_id":"120","session_id":"sess-1","timestamp":"2026-06-05T12:56:24.774753219Z","seq":3,"type":"run.token_delta","data":{"text":"Paris"}}`

	var ev HostedAgentEvent
	require.NoError(t, json.Unmarshal([]byte(frame), &ev))

	assert.Equal(t, "ev-9", ev.EventID)
	assert.Equal(t, "run-7", ev.RunID)
	assert.Equal(t, "sess-1", ev.SessionID)
	assert.Equal(t, uint64(3), ev.Seq)
	assert.Equal(t, HostedAgentEventKindTokenChunk, ev.Kind)
	assert.False(t, ev.At.IsZero())
	assert.JSONEq(t, `{"text":"Paris"}`, string(ev.Payload))

	assert.Empty(t, ev.SourceRaw, "absent source fields decode to zero, not an error")
	assert.Empty(t, ev.SourceEventID)
	assert.Empty(t, ev.SourceEventType)
}

// TestHostedAgentEvent_UnmarshalSPIWire_SourceFields pins the native-provenance
// fields: the runtime's own event id and type label, and the exact pre-mapping
// bytes (base64 on the wire).
func TestHostedAgentEvent_UnmarshalSPIWire_SourceFields(t *testing.T) {
	const native = `{"jsonrpc":"2.0","method":"item/agentMessage/delta","params":{"delta":"Paris"}}`

	frame := fmt.Sprintf(
		`{"event_id":"ev-9","run_id":"run-7","tenant_id":"120","session_id":"sess-1","timestamp":"2026-06-05T12:56:24Z","seq":3,"type":"run.token_delta","data":{"text":"Paris"},"source_event_id":"native-1","source_event_type":"item/agentMessage/delta","source_raw":%q}`,
		base64.StdEncoding.EncodeToString([]byte(native)),
	)

	var ev HostedAgentEvent
	require.NoError(t, json.Unmarshal([]byte(frame), &ev))

	assert.Equal(t, "native-1", ev.SourceEventID)
	assert.Equal(t, "item/agentMessage/delta", ev.SourceEventType)
	assert.Equal(t, native, string(ev.SourceRaw))
	// The canonical fields are unaffected by the presence of raw.
	assert.Equal(t, HostedAgentEventKindTokenChunk, ev.Kind)
	assert.JSONEq(t, `{"text":"Paris"}`, string(ev.Payload))
}

func TestHostedAgents_StreamSession_IncludeRaw(t *testing.T) {
	setup()
	defer teardown()

	const native = `{"jsonrpc":"2.0","method":"item/agentMessage/delta","params":{"delta":"hello"}}`

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "true", r.URL.Query().Get("include_raw"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ": connected\n\n")
		fmt.Fprintf(w,
			"id: ev-1\nevent: run.token_delta\ndata: {\"event_id\":\"ev-1\",\"run_id\":\"run-1\",\"tenant_id\":\"42\",\"session_id\":\"sess-abc123\",\"timestamp\":\"2026-03-01T12:01:00Z\",\"seq\":1,\"type\":\"run.token_delta\",\"data\":{\"text\":\"hello\"},\"source_event_type\":\"item/agentMessage/delta\",\"source_raw\":%q}\n\n",
			base64.StdEncoding.EncodeToString([]byte(native)),
		)
	})

	stream, _, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		IncludeRaw: true,
	})
	require.NoError(t, err)
	defer stream.Close()

	require.True(t, stream.Next())
	ev := stream.Current()
	assert.Equal(t, native, string(ev.SourceRaw))
	assert.Equal(t, "item/agentMessage/delta", ev.SourceEventType)
	assert.NoError(t, stream.Err())
}

// tenant_id is ignored, so an envelope carrying a non-numeric one still decodes.
func TestHostedAgentEvent_IgnoresTenantID(t *testing.T) {
	const frame = `{"event_id":"ev-3","run_id":"run-2","tenant_id":"not-a-number","session_id":"sess-1","timestamp":"2026-06-05T12:56:24Z","seq":4,"type":"run.token_delta","data":{"text":"hi"}}`

	var ev HostedAgentEvent
	require.NoError(t, json.Unmarshal([]byte(frame), &ev))

	assert.Equal(t, HostedAgentEventKindTokenChunk, ev.Kind)
	assert.JSONEq(t, `{"text":"hi"}`, string(ev.Payload))
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

func TestHostedAgents_ListSessions_ByName(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "godo-fixture", r.URL.Query().Get("name"))
		fmt.Fprintf(w, `{"sessions":[%s]}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.ListSessions(ctx, &HostedAgentSessionListOptions{
		Name: "godo-fixture",
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 1)
	assert.Equal(t, "godo-fixture", got.Sessions[0].Name)
	assert.Equal(t, hostedAgentSession, got.Sessions[0])
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

// TestHostedAgents_StreamSession_ReplayOnly pins the history lane on the same
// data-plane .../events endpoint as the live lane, selected by ?replay_only=true.
// The cursor moves to the replay_from query parameter here: on a history read it
// is an explicit pagination cursor, not the Last-Event-ID resume hint.
func TestHostedAgents_StreamSession_ReplayOnly(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "true", r.URL.Query().Get("replay_only"))
		assert.Equal(t, "ev-0", r.URL.Query().Get("replay_from"))
		assert.Empty(t, r.Header.Get("Last-Event-ID"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ": replay only\n\n")
	})

	stream, resp, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		ReplayFrom: "ev-0",
		ReplayOnly: true,
	})
	require.NoError(t, err)
	defer stream.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.False(t, stream.Next())
}

// TestHostedAgents_StreamSession_ReplayOnly_HistoryThenEOF covers what a
// replay-only read is for: the stored history arrives as ordinary events and the
// stream then ends on its own, which is what lets a history reader exit instead
// of blocking on a connection that stays open forever.
func TestHostedAgents_StreamSession_ReplayOnly_HistoryThenEOF(t *testing.T) {
	setup()
	defer teardown()

	const (
		catchingUp = `{"event_id":"","run_id":"","tenant_id":"0","session_id":"sess-abc123","timestamp":"2026-03-01T12:00:00Z","seq":0,"type":"stream.state","data":{"state":"catching_up","cursor":""}}`
		first      = `{"event_id":"ev-1","run_id":"run-1","tenant_id":"42","session_id":"sess-abc123","timestamp":"2026-03-01T12:01:00Z","seq":1,"type":"run.token_delta","data":{"text":"hello"}}`
		second     = `{"event_id":"ev-2","run_id":"run-1","tenant_id":"42","session_id":"sess-abc123","timestamp":"2026-03-01T12:02:00Z","seq":2,"type":"run.completed","data":{}}`
	)

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "true", r.URL.Query().Get("replay_only"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: stream.state\ndata: %s\n\n", catchingUp)
		fmt.Fprintf(w, "id: ev-1\nevent: run.token_delta\ndata: %s\n\n", first)
		fmt.Fprintf(w, "id: ev-2\nevent: run.completed\ndata: %s\n\n", second)
	})

	stream, _, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		ReplayOnly: true,
	})
	require.NoError(t, err)
	defer stream.Close()

	// A history read opens with catching_up and never reaches live.
	require.True(t, stream.Next())
	assert.Equal(t, HostedAgentEventKindStreamState, stream.Current().Kind)

	require.True(t, stream.Next())
	assert.Equal(t, HostedAgentEventKindTokenChunk, stream.Current().Kind)
	assert.Equal(t, "ev-1", stream.Current().EventID)

	require.True(t, stream.Next())
	assert.Equal(t, HostedAgentEventKindRunCompleted, stream.Current().Kind)
	assert.Equal(t, "ev-2", stream.Current().EventID)

	assert.False(t, stream.Next(), "replay-only stream must end after the last stored event")
	assert.NoError(t, stream.Err())
}

// A history page request carries before/limit alongside replay_only, and the
// trailing has_more comment is readable through HasMore once the page has been
// drained.
func TestHostedAgents_StreamSession_HistoryPage(t *testing.T) {
	setup()
	defer teardown()

	const eventJSON = `{"event_id":"ev-7","run_id":"run-1","tenant_id":"42","session_id":"sess-abc123","timestamp":"2026-03-01T12:01:00Z","seq":7,"type":"run.token_delta","data":{"text":"older"}}`

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		q := r.URL.Query()
		assert.Equal(t, "true", q.Get("replay_only"))
		assert.Equal(t, "ev-8", q.Get("before"))
		assert.Equal(t, "50", q.Get("limit"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ": connected to sess-abc123\n\n")
		fmt.Fprintf(w, "id: ev-7\nevent: run.token_delta\ndata: %s\n\n", eventJSON)
		fmt.Fprint(w, ": has_more=true\n\n")
	})

	stream, resp, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		ReplayOnly: true,
		Before:     "ev-8",
		Limit:      50,
	})
	require.NoError(t, err)
	defer stream.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	require.True(t, stream.Next())
	assert.Equal(t, "ev-7", stream.Current().EventID)
	// The trailer arrives after the last event, so HasMore only settles once
	// the page has been drained to its end.
	assert.False(t, stream.Next())
	assert.NoError(t, stream.Err())
	assert.True(t, stream.HasMore())
}

// The oldest page reports has_more=false, ending a backward walk.
func TestHostedAgents_StreamSession_HistoryPageHasMoreFalse(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ": has_more=false\n\n")
	})

	stream, _, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		ReplayOnly: true,
		Before:     "ev-1",
	})
	require.NoError(t, err)
	defer stream.Close()

	assert.False(t, stream.Next())
	assert.False(t, stream.HasMore())
}

// A cursorless replay omits before/limit entirely, and reports no has_more.
func TestHostedAgents_StreamSession_NoPagingParamsWhenUnset(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/events", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.False(t, q.Has("before"))
		assert.False(t, q.Has("limit"))
		assert.False(t, q.Has("include_raw"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ": connected\n\n")
	})

	stream, _, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		ReplayOnly: true,
	})
	require.NoError(t, err)
	defer stream.Close()

	assert.False(t, stream.Next())
	assert.False(t, stream.HasMore())
}

// TestHostedAgents_StreamSession_StreamState pins the data plane's transport
// control frame. It arrives in the same canonical envelope as an agent event, so
// it decodes through the same parser: callers identify it by Kind and skip it
// rather than rendering it as session activity.
func TestHostedAgents_StreamSession_StreamState(t *testing.T) {
	setup()
	defer teardown()

	const frame = `{"event_id":"","run_id":"","tenant_id":"0","session_id":"sess-abc123","timestamp":"2026-03-01T12:00:00Z","seq":0,"type":"stream.state","data":{"state":"live","cursor":""}}`

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: stream.state\ndata: %s\n\n", frame)
	})

	stream, _, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", nil)
	require.NoError(t, err)
	defer stream.Close()

	require.True(t, stream.Next())
	ev := stream.Current()
	assert.Equal(t, HostedAgentEventKindStreamState, ev.Kind)
	assert.Equal(t, "sess-abc123", ev.SessionID)
	// A control frame is not a durable event, so it carries no event_id of its
	// own and the server sends no `id:` line for it. (Per the SSE spec the
	// reader's last-event-id buffer persists, so a control frame arriving after
	// an event reports that event's id — never a new cursor position.)
	assert.Empty(t, ev.EventID)

	var state HostedAgentStreamState
	require.NoError(t, json.Unmarshal(ev.Payload, &state))
	assert.Equal(t, HostedAgentStreamStateLive, state.State)
	assert.Empty(t, state.Cursor)
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

	_, err = client.HostedAgents.PauseSession(ctx, "")
	require.EqualError(t, err, "hosted agents: session id is required")

	_, err = client.HostedAgents.ResumeSession(ctx, "")
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

	_, _, err = client.HostedAgents.RelayRequest(ctx, "", &HostedAgentRelayRequest{
		SourceRaw: []byte(`{}`),
	})
	require.EqualError(t, err, "hosted agents: session id is required")

	_, _, err = client.HostedAgents.RelayRequest(ctx, "sess-abc123", nil)
	require.EqualError(t, err, "hosted agents: source_raw is required")

	// A relay with nothing to relay would block on the agent for no reason.
	_, _, err = client.HostedAgents.RelayRequest(ctx, "sess-abc123", &HostedAgentRelayRequest{})
	require.EqualError(t, err, "hosted agents: source_raw is required")

	_, _, err = client.HostedAgents.ExecInSandbox(ctx, "sess-abc123", &HostedAgentSandboxExecRequest{})
	require.EqualError(t, err, "hosted agents: argv is required")

	// A live attach cannot start in the past; the server answers this with a
	// 400, so it is rejected before the request goes out.
	_, _, err = client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		Before: "ev-1",
	})
	require.EqualError(t, err, "hosted agents: before requires replay only")
}

func TestHostedAgents_UploadWorkspace(t *testing.T) {
	setup()
	defer teardown()

	const payload = "hello world"

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/upload", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "application/octet-stream", r.Header.Get("Content-Type"))
		assert.Equal(t, "src/main.go", r.URL.Query().Get("path"))
		assert.Equal(t, "true", r.URL.Query().Get("is_archive"))
		assert.Equal(t, "deadbeef", r.Header.Get("X-Content-Sha256"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, payload, string(body))
		fmt.Fprintf(w, `{"path":"/workspace/src/main.go","bytes_written":%d}`, len(payload))
	})

	got, resp, err := client.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{
		Path:          "src/main.go",
		IsArchive:     true,
		ContentSHA256: "deadbeef",
		Body:          strings.NewReader(payload),
	})
	require.NoError(t, err)
	assert.Equal(t, "/workspace/src/main.go", got.Path)
	assert.Equal(t, int64(len(payload)), got.BytesWritten)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// workspaceUploadFile writes n bytes to a temp file and returns it positioned
// at the start, mirroring how doctl hands a file to UploadWorkspace.
func workspaceUploadFile(t *testing.T, n int) *os.File {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "upload")
	require.NoError(t, err)
	t.Cleanup(func() { f.Close() })

	_, err = f.Write(bytes.Repeat([]byte("x"), n))
	require.NoError(t, err)
	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err)

	return f
}

func TestHostedAgents_UploadWorkspace_DeclaresContentLength(t *testing.T) {
	const payload = "hello world"

	tests := []struct {
		name string
		body func(t *testing.T) io.Reader
		// contentLength is the explicit request field, left zero to exercise
		// inference from the body.
		contentLength int64
		want          int64
	}{
		{
			name: "os.File is measured by stat",
			body: func(t *testing.T) io.Reader { return workspaceUploadFile(t, len(payload)) },
			want: int64(len(payload)),
		},
		{
			name: "os.File is measured from its current offset",
			body: func(t *testing.T) io.Reader {
				f := workspaceUploadFile(t, len(payload))
				_, err := f.Seek(4, io.SeekStart)
				require.NoError(t, err)
				return f
			},
			want: int64(len(payload) - 4),
		},
		{
			name: "in-memory reader reports its remainder",
			body: func(t *testing.T) io.Reader { return strings.NewReader(payload) },
			want: int64(len(payload)),
		},
		{
			name: "opaque reader falls back to the explicit length",
			body: func(t *testing.T) io.Reader {
				// Hides Len() so neither http.NewRequest nor uploadBodyLength
				// can measure it.
				return struct{ io.Reader }{strings.NewReader(payload)}
			},
			contentLength: int64(len(payload)),
			want:          int64(len(payload)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			defer teardown()

			var (
				gotLength   int64
				gotEncoding []string
			)
			mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/upload", func(w http.ResponseWriter, r *http.Request) {
				gotLength = r.ContentLength
				gotEncoding = r.TransferEncoding
				io.Copy(io.Discard, r.Body)
				fmt.Fprint(w, `{"path":"/workspace/f","bytes_written":0}`)
			})

			_, _, err := client.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{
				Path:          "f",
				Body:          tt.body(t),
				ContentLength: tt.contentLength,
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, gotLength, "declared Content-Length")
			assert.Empty(t, gotEncoding, "payload must not be chunked")
		})
	}
}

func TestHostedAgents_UploadWorkspace_ExpectContinueThreshold(t *testing.T) {
	tests := []struct {
		name string
		size int
		want string
	}{
		{name: "below threshold negotiates nothing", size: 16, want: ""},
		{name: "at threshold negotiates", size: workspaceUploadExpectContinueMinBytes, want: "100-continue"},
		{name: "above threshold negotiates", size: workspaceUploadExpectContinueMinBytes + 1, want: "100-continue"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			defer teardown()

			var got string
			mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/upload", func(w http.ResponseWriter, r *http.Request) {
				got = r.Header.Get("Expect")
				io.Copy(io.Discard, r.Body)
				fmt.Fprint(w, `{"path":"/workspace/f","bytes_written":0}`)
			})

			_, _, err := client.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{
				Path: "f",
				Body: workspaceUploadFile(t, tt.size),
			})
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// countingListener totals the bytes read off every accepted connection, which
// is how the tests below distinguish "payload withheld" from "payload sent".
type countingListener struct {
	net.Listener
	read *int64
}

func (l countingListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return countingConn{Conn: c, read: l.read}, nil
}

type countingConn struct {
	net.Conn
	read *int64
}

func (c countingConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	atomic.AddInt64(c.read, int64(n))
	return n, err
}

// newCountingUploadServer serves handler and reports bytes received on the wire.
func newCountingUploadServer(t *testing.T, handler http.HandlerFunc) (*Client, *int64) {
	t.Helper()

	var read int64
	srv := httptest.NewUnstartedServer(handler)
	srv.Listener = countingListener{Listener: srv.Listener, read: &read}
	srv.Start()
	t.Cleanup(srv.Close)

	// NewClient does not retry, so a 503 surfaces on the first attempt rather
	// than being replayed.
	c := NewClient(nil)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	c.BaseURL = u

	return c, &read
}

// The point of the Expect handshake: when OHS refuses from the headers alone --
// its size cap or an exhausted transfer slot -- the payload never leaves the
// client.
func TestHostedAgents_UploadWorkspace_WithholdsPayloadWhenRefusedAtHeaders(t *testing.T) {
	const size = 4 << 20

	var gotExpect string
	c, read := newCountingUploadServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotExpect = r.Header.Get("Expect")
		// Refuse without touching r.Body, so net/http never emits 100 Continue.
		w.Header().Set("Retry-After", "1")
		http.Error(w, "workspace transfer capacity exhausted, retry later", http.StatusServiceUnavailable)
	})

	_, resp, err := c.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{
		Path: "big.bin",
		Body: workspaceUploadFile(t, size),
	})

	require.Error(t, err, "a 503 must surface to the caller")
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, "100-continue", gotExpect)
	assert.Equal(t, "1", resp.Header.Get("Retry-After"))

	// Only the request headers should have crossed the wire.
	assert.Less(t, atomic.LoadInt64(read), int64(32<<10),
		"payload was transmitted despite the request being refused at headers")
}

// HTTP/2 implements the handshake in a completely separate transport, and the
// API is reached over TLS with h2 negotiation enabled, so the guarantee has to
// hold there too.
func TestHostedAgents_UploadWorkspace_WithholdsPayloadOverHTTP2(t *testing.T) {
	const size = 4 << 20

	var (
		read     int64
		gotProto string
	)
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotProto = r.Proto
		http.Error(w, "workspace transfer capacity exhausted, retry later", http.StatusServiceUnavailable)
	}))
	srv.Listener = countingListener{Listener: srv.Listener, read: &read}
	srv.EnableHTTP2 = true
	srv.StartTLS()
	defer srv.Close()

	hc := srv.Client()
	// httptest leaves ExpectContinueTimeout at zero, and the h2 transport reads
	// it to decide whether to negotiate at all -- zero means send the body
	// immediately. Every transport godo runs on in production sets it (both
	// http.DefaultTransport and cleanhttp use 1s), so match them or this would
	// test a configuration that does not exist.
	hc.Transport.(*http.Transport).ExpectContinueTimeout = time.Second

	c := NewClient(hc)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	c.BaseURL = u

	_, resp, err := c.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{
		Path: "big.bin",
		Body: workspaceUploadFile(t, size),
	})

	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	require.Equal(t, "HTTP/2.0", gotProto, "test did not exercise h2")

	// The handler is deliberately not asserted on Expect: net/http lifts it out
	// of the headers into an internal flag before an h2 handler runs, so the
	// negotiation is invisible server-side and only the byte count can show it
	// happened.
	//
	// Counted before TLS decryption, so the bound allows for handshake and
	// framing overhead while staying far below the payload.
	assert.Less(t, atomic.LoadInt64(&read), int64(64<<10),
		"payload was transmitted over h2 despite being refused at headers")
}

// The withholding above must survive the client stack doctl actually builds:
// an oauth2 client wrapped by the retryablehttp transport. That transport is
// where Expect could quietly stop working, since honouring it depends on a
// non-zero ExpectContinueTimeout, and where a retry could replay the payload.
func TestHostedAgents_UploadWorkspace_WithholdsPayloadThroughRetryTransport(t *testing.T) {
	const size = 4 << 20

	var (
		attempts int64
		read     int64
	)
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&attempts, 1)
		assert.Equal(t, "100-continue", r.Header.Get("Expect"))
		w.Header().Set("Retry-After", "1")
		http.Error(w, "workspace transfer capacity exhausted, retry later", http.StatusServiceUnavailable)
	}))
	srv.Listener = countingListener{Listener: srv.Listener, read: &read}
	srv.Start()
	defer srv.Close()

	// Mirrors doctl: oauth2 client + WithRetryAndBackoffs. The waits are
	// squeezed only to keep the test quick.
	oauthClient := oauth2.NewClient(context.Background(),
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	c, err := New(oauthClient,
		WithRetryAndBackoffs(RetryConfig{
			RetryMax:     1,
			RetryWaitMin: PtrTo(0.01),
			RetryWaitMax: PtrTo(0.02),
		}),
		SetBaseURL(srv.URL),
	)
	require.NoError(t, err)

	start := time.Now()
	_, resp, err := c.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{
		Path: "big.bin",
		Body: workspaceUploadFile(t, size),
	})
	elapsed := time.Since(start)

	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	// A 503 is retryable, so the shed must actually be replayed rather than
	// surfacing on the first attempt.
	assert.Equal(t, int64(2), atomic.LoadInt64(&attempts), "initial attempt plus one retry")

	// The retry must wait out the Retry-After OHS sent rather than the far
	// shorter backoff configured above, so a shedding server is not hammered.
	assert.GreaterOrEqual(t, elapsed, time.Second, "Retry-After was ignored")

	// Neither attempt may put the payload on the wire.
	assert.Less(t, atomic.LoadInt64(&read), int64(32<<10),
		"payload was transmitted through the retry transport despite being refused at headers")
}

// Control for the test above: an accepted upload must still deliver every byte.
func TestHostedAgents_UploadWorkspace_SendsPayloadWhenAccepted(t *testing.T) {
	const size = 4 << 20

	var got int64
	c, read := newCountingUploadServer(t, func(w http.ResponseWriter, r *http.Request) {
		n, err := io.Copy(io.Discard, r.Body)
		require.NoError(t, err)
		got = n
		fmt.Fprintf(w, `{"path":"/workspace/big.bin","bytes_written":%d}`, n)
	})

	out, resp, err := c.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{
		Path: "big.bin",
		Body: workspaceUploadFile(t, size),
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int64(size), got, "handler must receive the whole payload")
	assert.Equal(t, int64(size), out.BytesWritten)
	assert.GreaterOrEqual(t, atomic.LoadInt64(read), int64(size))
}

func TestUploadBodyLength(t *testing.T) {
	t.Run("explicit length wins over an inferrable body", func(t *testing.T) {
		got := uploadBodyLength(&HostedAgentWorkspaceUploadRequest{
			Body:          workspaceUploadFile(t, 4096),
			ContentLength: 7,
		})
		assert.Equal(t, int64(7), got)
	})

	t.Run("non-regular file is not measurable", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		t.Cleanup(func() { r.Close(); w.Close() })

		assert.Zero(t, uploadBodyLength(&HostedAgentWorkspaceUploadRequest{Body: r}))
	})

	t.Run("opaque reader is not measurable", func(t *testing.T) {
		body := struct{ io.Reader }{strings.NewReader("hello")}
		assert.Zero(t, uploadBodyLength(&HostedAgentWorkspaceUploadRequest{Body: body}))
	})

	t.Run("fully consumed file measures zero", func(t *testing.T) {
		f := workspaceUploadFile(t, 8)
		_, err := io.Copy(io.Discard, f)
		require.NoError(t, err)

		assert.Zero(t, uploadBodyLength(&HostedAgentWorkspaceUploadRequest{Body: f}))
	})
}

func workspaceDownloadFooter(payload string) string {
	sum := sha256.Sum256([]byte(payload))
	return workspaceDownloadFooterPrefix + hex.EncodeToString(sum[:]) + "\n"
}

func TestHostedAgents_DownloadWorkspace(t *testing.T) {
	setup()
	defer teardown()

	const payload = "the quick brown fox"
	footer := workspaceDownloadFooter(payload)

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "notes.txt", r.URL.Query().Get("path"))
		assert.Equal(t, "true", r.URL.Query().Get("as_archive"))

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Workspace-Is-Archive", "true")
		w.Header().Set("X-Workspace-Size-Bytes", strconv.Itoa(len(payload)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload + footer))
	})

	dl, resp, err := client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{
		Path:      "notes.txt",
		AsArchive: true,
	})
	require.NoError(t, err)
	require.NotNil(t, dl)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, dl.IsArchive)
	assert.Equal(t, int64(len(payload)), dl.SizeBytes)

	body, err := io.ReadAll(dl.Body)
	require.NoError(t, err)
	assert.Equal(t, payload, string(body))
	require.NoError(t, dl.Body.Close())
}

func TestHostedAgents_DownloadWorkspace_ChecksumMismatch(t *testing.T) {
	setup()
	defer teardown()

	const payload = "the quick brown fox"
	badFooter := workspaceDownloadFooterPrefix + strings.Repeat("0", 64) + "\n"

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload + badFooter))
	})

	dl, _, err := client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{
		Path: "notes.txt",
	})
	require.NoError(t, err)
	defer dl.Body.Close()

	_, err = io.ReadAll(dl.Body)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestHostedAgents_DownloadWorkspace_MissingFooter(t *testing.T) {
	setup()
	defer teardown()

	const payload = "partial data"

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload))
	})

	dl, _, err := client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{
		Path: "notes.txt",
	})
	require.NoError(t, err)
	defer dl.Body.Close()

	_, err = io.ReadAll(dl.Body)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "integrity footer")
}

func TestHostedAgents_DownloadWorkspace_InvalidFooter(t *testing.T) {
	setup()
	defer teardown()

	const payload = "the quick brown fox"
	invalidFooter := "NOTASHA1" + strings.Repeat("a", 64) + "\n"

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload + invalidFooter))
	})

	dl, _, err := client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{
		Path: "notes.txt",
	})
	require.NoError(t, err)
	defer dl.Body.Close()

	_, err = io.ReadAll(dl.Body)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid workspace download integrity footer")
}

func TestHostedAgents_DownloadWorkspace_EmptyPayload(t *testing.T) {
	setup()
	defer teardown()

	const payload = ""
	footer := workspaceDownloadFooter(payload)

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Workspace-Size-Bytes", "0")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(footer))
	})

	dl, _, err := client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{
		Path: "empty.txt",
	})
	require.NoError(t, err)
	defer dl.Body.Close()

	body, err := io.ReadAll(dl.Body)
	require.NoError(t, err)
	assert.Equal(t, "", string(body))
	assert.Equal(t, int64(0), dl.SizeBytes)
}

func TestHostedAgents_WorkspaceValidationErrors(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.UploadWorkspace(ctx, "", &HostedAgentWorkspaceUploadRequest{})
	require.EqualError(t, err, "hosted agents: session id is required")

	_, _, err = client.HostedAgents.UploadWorkspace(ctx, "sess-abc123", nil)
	require.EqualError(t, err, "hosted agents: upload request is required")

	_, _, err = client.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{})
	require.EqualError(t, err, "hosted agents: path is required")

	_, _, err = client.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{Path: "x"})
	require.EqualError(t, err, "hosted agents: body is required")

	_, _, err = client.HostedAgents.DownloadWorkspace(ctx, "", &HostedAgentWorkspaceDownloadRequest{})
	require.EqualError(t, err, "hosted agents: session id is required")

	_, _, err = client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", nil)
	require.EqualError(t, err, "hosted agents: download request is required")

	_, _, err = client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{})
	require.EqualError(t, err, "hosted agents: path is required")
}

func TestHostedAgents_CreateWorkspaceTransfer_Upload(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got HostedAgentWorkspaceTransferCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, HostedAgentWorkspaceTransferDirectionUpload, got.Direction)
		assert.Equal(t, "/workspace/data/big.bin", got.Path)
		assert.Equal(t, int64(524288000), got.SizeBytes)
		assert.Equal(t, "abc123", got.SHA256)
		assert.False(t, got.IsArchive)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{
			"transfer_id": "xfer-upload-1",
			"direction": "upload",
			"status": "pending",
			"upload_id": "up-1",
			"part_size": 16777216,
			"expires_at": "2026-07-21T12:00:00Z"
		}`)
	})

	got, resp, err := client.HostedAgents.CreateWorkspaceTransfer(ctx, "sess-abc123", &HostedAgentWorkspaceTransferCreateRequest{
		Direction: HostedAgentWorkspaceTransferDirectionUpload,
		Path:      "/workspace/data/big.bin",
		SizeBytes: 524288000,
		SHA256:    "abc123",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "xfer-upload-1", got.TransferID)
	assert.Equal(t, HostedAgentWorkspaceTransferDirectionUpload, got.Direction)
	assert.Equal(t, HostedAgentWorkspaceTransferStatusPending, got.Status)
	assert.Equal(t, "up-1", got.UploadID)
	assert.Equal(t, int64(16777216), got.PartSize)
	require.NotNil(t, got.ExpiresAt)
	assert.Equal(t, time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC), got.ExpiresAt.Time)
}

func TestHostedAgents_CreateWorkspaceTransfer_Download(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got HostedAgentWorkspaceTransferCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, HostedAgentWorkspaceTransferDirectionDownload, got.Direction)
		assert.Equal(t, "/workspace/data/big.bin", got.Path)
		assert.True(t, got.AsArchive)

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{
			"transfer_id": "xfer-download-1",
			"direction": "download",
			"status": "pending"
		}`)
	})

	got, resp, err := client.HostedAgents.CreateWorkspaceTransfer(ctx, "sess-abc123", &HostedAgentWorkspaceTransferCreateRequest{
		Direction: HostedAgentWorkspaceTransferDirectionDownload,
		Path:      "/workspace/data/big.bin",
		AsArchive: true,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, "xfer-download-1", got.TransferID)
	assert.Equal(t, HostedAgentWorkspaceTransferStatusPending, got.Status)
}

func TestHostedAgents_CreateWorkspaceTransferPartUploadURLs(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers/xfer-1/part-upload-urls", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got HostedAgentWorkspaceTransferPartUploadURLsRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, []int{1}, got.PartNumbers)

		fmt.Fprint(w, `{
			"part_urls": [
				{"part_number": 1, "upload_url": "https://spaces.example/part-1"}
			]
		}`)
	})

	got, resp, err := client.HostedAgents.CreateWorkspaceTransferPartUploadURLs(ctx, "sess-abc123", "xfer-1", &HostedAgentWorkspaceTransferPartUploadURLsRequest{
		PartNumbers: []int{1},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, got.PartURLs, 1)
	assert.Equal(t, 1, got.PartURLs[0].PartNumber)
	assert.Equal(t, "https://spaces.example/part-1", got.PartURLs[0].UploadURL)
}

func TestHostedAgents_CommitWorkspaceTransfer(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers/xfer-1/commit", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got HostedAgentWorkspaceTransferCommitRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "deadbeef", got.SHA256)

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{
			"transfer_id": "xfer-1",
			"status": "in_progress",
			"size_bytes": 524288000
		}`)
	})

	got, resp, err := client.HostedAgents.CommitWorkspaceTransfer(ctx, "sess-abc123", "xfer-1", &HostedAgentWorkspaceTransferCommitRequest{
		SHA256: "deadbeef",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, HostedAgentWorkspaceTransferStatusInProgress, got.Status)
	assert.Equal(t, int64(524288000), got.SizeBytes)
}

func TestHostedAgents_GetWorkspaceTransfer(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers/xfer-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"transfer_id": "xfer-1",
			"direction": "download",
			"status": "completed",
			"bytes_written": 524288000,
			"sha256": "e6b84c4839cbbfc3cde3b1bc84e8a82b9661c00eae1726fcff0dca8d643423ae",
			"download_url": "https://spaces.example/download",
			"expires_at": "2026-07-21T13:00:00Z",
			"error_message": null
		}`)
	})

	got, resp, err := client.HostedAgents.GetWorkspaceTransfer(ctx, "sess-abc123", "xfer-1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, HostedAgentWorkspaceTransferDirectionDownload, got.Direction)
	assert.Equal(t, HostedAgentWorkspaceTransferStatusCompleted, got.Status)
	assert.Equal(t, int64(524288000), got.BytesWritten)
	assert.Equal(t, "e6b84c4839cbbfc3cde3b1bc84e8a82b9661c00eae1726fcff0dca8d643423ae", got.SHA256)
	assert.Equal(t, "https://spaces.example/download", got.DownloadURL)
	require.NotNil(t, got.ExpiresAt)
	assert.Equal(t, time.Date(2026, 7, 21, 13, 0, 0, 0, time.UTC), got.ExpiresAt.Time)
	assert.Empty(t, got.ErrorMessage)
}

func TestHostedAgents_CancelWorkspaceTransfer(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers/xfer-1/cancel", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got HostedAgentWorkspaceTransferCancelRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "user cancelled", got.Reason)

		fmt.Fprint(w, `{
			"transfer_id": "xfer-1",
			"aborted": true,
			"status": "failed"
		}`)
	})

	got, resp, err := client.HostedAgents.CancelWorkspaceTransfer(ctx, "sess-abc123", "xfer-1", &HostedAgentWorkspaceTransferCancelRequest{
		Reason: "user cancelled",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "xfer-1", got.TransferID)
	assert.True(t, got.Aborted)
	assert.Equal(t, HostedAgentWorkspaceTransferStatusFailed, got.Status)
}

func TestHostedAgents_WorkspaceTransferValidationErrors(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.CreateWorkspaceTransfer(ctx, "", &HostedAgentWorkspaceTransferCreateRequest{})
	require.EqualError(t, err, "hosted agents: session id is required")

	_, _, err = client.HostedAgents.CreateWorkspaceTransfer(ctx, "sess-abc123", nil)
	require.EqualError(t, err, "hosted agents: transfer create request is required")

	_, _, err = client.HostedAgents.CreateWorkspaceTransfer(ctx, "sess-abc123", &HostedAgentWorkspaceTransferCreateRequest{})
	require.EqualError(t, err, "hosted agents: direction is required")

	_, _, err = client.HostedAgents.CreateWorkspaceTransfer(ctx, "sess-abc123", &HostedAgentWorkspaceTransferCreateRequest{
		Direction: HostedAgentWorkspaceTransferDirectionUpload,
	})
	require.EqualError(t, err, "hosted agents: path is required")

	_, _, err = client.HostedAgents.CreateWorkspaceTransferPartUploadURLs(ctx, "sess-abc123", "xfer-1", &HostedAgentWorkspaceTransferPartUploadURLsRequest{})
	require.EqualError(t, err, "hosted agents: part_numbers must not be empty")

	_, _, err = client.HostedAgents.GetWorkspaceTransfer(ctx, "sess-abc123", "")
	require.EqualError(t, err, "hosted agents: transfer id is required")

	_, _, err = client.HostedAgents.CancelWorkspaceTransfer(ctx, "", "xfer-1", nil)
	require.EqualError(t, err, "hosted agents: session id is required")
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

func TestHostedAgents_StartProviderAuth(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/auth/github", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"provider": "github",
			"status": "pending",
			"connect_url": "https://cloud.digitalocean.com/security/connectlinks/confirm?token=abc",
			"poll_url": "https://cloud.digitalocean.com/api/v1/security/connectlinks/poll?token=def",
			"verification_code": "k5r2cprq",
			"expires_at": "2036-08-10T10:44:32Z"
		}`)
	})

	got, resp, err := client.HostedAgents.StartProviderAuth(ctx, "github")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &HostedAgentProviderAuthStart{
		Provider:         "github",
		Status:           "pending",
		ConnectURL:       "https://cloud.digitalocean.com/security/connectlinks/confirm?token=abc",
		PollURL:          "https://cloud.digitalocean.com/api/v1/security/connectlinks/poll?token=def",
		VerificationCode: "k5r2cprq",
		ExpiresAt:        &Timestamp{Time: time.Date(2036, 8, 10, 10, 44, 32, 0, time.UTC)},
	}, got)
}

func TestHostedAgents_StartProviderAuth_AlreadyConnected(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/auth/github", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"provider":"github","status":"success"}`)
	})

	got, _, err := client.HostedAgents.StartProviderAuth(ctx, "github")
	require.NoError(t, err)
	assert.Equal(t, "success", got.Status)
	assert.Empty(t, got.ConnectURL)
	assert.Empty(t, got.PollURL)
}

func TestHostedAgents_PollProviderAuth(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/auth/github/poll", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "https://cloud.digitalocean.com/api/v1/security/connectlinks/poll?token=def", r.URL.Query().Get("poll_url"))
		fmt.Fprint(w, `{"provider":"github","status":"success","expires_at":"2036-08-10T10:44:32Z"}`)
	})

	got, resp, err := client.HostedAgents.PollProviderAuth(ctx, "github", "https://cloud.digitalocean.com/api/v1/security/connectlinks/poll?token=def")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &HostedAgentProviderAuthPoll{
		Provider:  "github",
		Status:    "success",
		ExpiresAt: &Timestamp{Time: time.Date(2036, 8, 10, 10, 44, 32, 0, time.UTC)},
	}, got)
}

func TestHostedAgents_ProviderAuth_RequiredArgs(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.StartProviderAuth(ctx, "")
	require.EqualError(t, err, "hosted agents: provider is required")

	_, _, err = client.HostedAgents.PollProviderAuth(ctx, "", "https://example.com/poll")
	require.EqualError(t, err, "hosted agents: provider is required")

	_, _, err = client.HostedAgents.PollProviderAuth(ctx, "github", "")
	require.EqualError(t, err, "hosted agents: poll url is required")
}
