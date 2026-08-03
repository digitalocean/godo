package godo

import (
	"bytes"
	"crypto/sha256"
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
)

var hostedAgentSession = HostedAgentSession{
	SessionID:   "sess-abc123",
	Name:        "godo-fixture",
	TeamID:      42,
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
	"team_id": 42,
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
				"team_id": 42,
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
				"team_id": 7,
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
				"team_id": 9,
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
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "agents.digitalocean.com/v1alpha1")
		assert.Contains(t, string(body), "adapter: opencode")
		fmt.Fprintf(w, `{"session":%s}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.CreateSessionFromManifest(ctx, []byte(manifest))
	require.NoError(t, err)
	assert.Equal(t, hostedAgentSession, *got)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_CreateSessionFromManifest_Empty(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.CreateSessionFromManifest(ctx, []byte("  \n"))
	require.EqualError(t, err, "hosted agents: manifest is required")
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

func TestHostedAgents_StreamSession(t *testing.T) {
	setup()
	defer teardown()

	// The server serializes the SPI canonical event envelope: the discriminator
	// is `type` (dot-separated), the body is `data`, the timestamp is
	// `timestamp`, and the team id rides as a decimal string in `tenant_id`.
	const eventJSON = `{"event_id":"ev-1","run_id":"run-1","tenant_id":"42","session_id":"sess-abc123","timestamp":"2026-03-01T12:01:00Z","seq":1,"type":"run.token_delta","data":{"text":"hello"}}`

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/stream", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))
		assert.Equal(t, "ev-0", r.URL.Query().Get("replay_from"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, ": connected\n\n")
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
	assert.Equal(t, uint64(42), ev.TeamID)
	assert.JSONEq(t, `{"text":"hello"}`, string(ev.Payload))
	assert.NoError(t, stream.Err())
	assert.False(t, stream.Next())
}

// TestHostedAgentEvent_UnmarshalSPIWire pins the SPI canonical envelope decode:
// type->Kind (dot-separated), data->Payload, timestamp->At, tenant_id(string)->TeamID.
func TestHostedAgentEvent_UnmarshalSPIWire(t *testing.T) {
	const frame = `{"event_id":"ev-9","run_id":"run-7","tenant_id":"120","session_id":"sess-1","timestamp":"2026-06-05T12:56:24.774753219Z","seq":3,"type":"run.token_delta","data":{"text":"Paris"}}`

	var ev HostedAgentEvent
	require.NoError(t, json.Unmarshal([]byte(frame), &ev))

	assert.Equal(t, "ev-9", ev.EventID)
	assert.Equal(t, "run-7", ev.RunID)
	assert.Equal(t, "sess-1", ev.SessionID)
	assert.Equal(t, uint64(120), ev.TeamID)
	assert.Equal(t, uint64(3), ev.Seq)
	assert.Equal(t, HostedAgentEventKindTokenChunk, ev.Kind)
	assert.False(t, ev.At.IsZero())
	assert.JSONEq(t, `{"text":"Paris"}`, string(ev.Payload))
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

	_, _, err = client.HostedAgents.ExecInSandbox(ctx, "sess-abc123", &HostedAgentSandboxExecRequest{})
	require.EqualError(t, err, "hosted agents: argv is required")
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
