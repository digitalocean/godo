package godo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const hostedAgentWorkspaceTestID = "019fb39c-14d9-7080-933e-b9b90e25acda"

var hostedAgentWorkspace = HostedAgentWorkspace{
	WorkspaceID:       hostedAgentWorkspaceTestID,
	Name:              "scratch",
	State:             HostedAgentWorkspaceStateAttached,
	AttachedSessionID: "sess-abc123",
	SizeGibibytes:     20,
	BytesUsed:         734003200,
	LastSavedAt:       &Timestamp{Time: time.Date(2026, 10, 5, 9, 30, 0, 0, time.UTC)},
	CreatedAt:         Timestamp{Time: time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)},
	UpdatedAt:         Timestamp{Time: time.Date(2026, 10, 5, 9, 31, 0, 0, time.UTC)},
}

var hostedAgentWorkspaceJSON = `
{
	"workspace_id": "019fb39c-14d9-7080-933e-b9b90e25acda",
	"name": "scratch",
	"state": "ATTACHED",
	"attached_session_id": "sess-abc123",
	"size_gibibytes": 20,
	"bytes_used": 734003200,
	"last_saved_at": "2026-10-05T09:30:00Z",
	"created_at": "2026-10-01T08:00:00Z",
	"updated_at": "2026-10-05T09:31:00Z"
}
`

func TestHostedAgents_CreateWorkspace(t *testing.T) {
	// The API answers 201 for a new workspace and 200 for an idempotent replay.
	for _, status := range []int{http.StatusCreated, http.StatusOK} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc("/v2/agents/workspaces", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodPost)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				raw, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.JSONEq(t, `{"size_gibibytes":20,"name":"scratch"}`, string(raw))
				w.WriteHeader(status)
				fmt.Fprintf(w, `{"workspace":%s}`, hostedAgentWorkspaceJSON)
			})

			got, resp, err := client.HostedAgents.CreateWorkspace(ctx, &HostedAgentWorkspaceCreateRequest{
				SizeGibibytes: 20,
				Name:          "scratch",
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, status, resp.StatusCode)
			assert.Equal(t, hostedAgentWorkspace, *got)
		})
	}
}

func TestHostedAgents_CreateWorkspace_OmitsEmptyName(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/workspaces", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"size_gibibytes":5}`, string(raw))
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"workspace":%s}`, hostedAgentWorkspaceJSON)
	})

	_, _, err := client.HostedAgents.CreateWorkspace(ctx, &HostedAgentWorkspaceCreateRequest{SizeGibibytes: 5})
	require.NoError(t, err)
}

func TestHostedAgents_CreateWorkspace_IdempotencyKey(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		wantHeader bool
	}{
		{name: "key set", key: "create-scratch-1", wantHeader: true},
		{name: "key empty", key: "", wantHeader: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc("/v2/agents/workspaces", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodPost)
				values, present := r.Header["Idempotency-Key"]
				assert.Equal(t, tt.wantHeader, present)
				if tt.wantHeader {
					assert.Equal(t, []string{tt.key}, values)
				}

				// The key travels as a header only, never in the body.
				raw, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				var fields map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(raw, &fields))
				assert.Contains(t, fields, "size_gibibytes")
				assert.Contains(t, fields, "name")
				assert.NotContains(t, fields, "idempotency_key")
				assert.NotContains(t, fields, "IdempotencyKey")
				if tt.key != "" {
					assert.NotContains(t, string(raw), tt.key)
				}

				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, `{"workspace":%s}`, hostedAgentWorkspaceJSON)
			})

			got, _, err := client.HostedAgents.CreateWorkspace(ctx, &HostedAgentWorkspaceCreateRequest{
				SizeGibibytes:  20,
				Name:           "scratch",
				IdempotencyKey: tt.key,
			})
			require.NoError(t, err)
			assert.Equal(t, hostedAgentWorkspaceTestID, got.WorkspaceID)
		})
	}
}

func TestHostedAgents_CreateWorkspace_NilRequest(t *testing.T) {
	got, resp, err := client.HostedAgents.CreateWorkspace(ctx, nil)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Nil(t, resp)
}

func TestHostedAgents_CreateWorkspace_NoWorkspaceInResponse(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/workspaces", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{}`)
	})

	got, _, err := client.HostedAgents.CreateWorkspace(ctx, &HostedAgentWorkspaceCreateRequest{SizeGibibytes: 20})
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestHostedAgents_CreateWorkspace_APIErrors(t *testing.T) {
	tests := []struct {
		status  int
		message string
	}{
		{status: http.StatusBadRequest, message: "size_gibibytes must be between 1 and 100"},
		{status: http.StatusConflict, message: "workspace limit reached"},
		{status: http.StatusUnprocessableEntity, message: "idempotency key reused with a different request"},
		{status: http.StatusNotImplemented, message: "workspaces are not enabled"},
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc("/v2/agents/workspaces", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodPost)
				w.WriteHeader(tt.status)
				fmt.Fprintf(w, `{"id":"err","message":%q}`, tt.message)
			})

			got, resp, err := client.HostedAgents.CreateWorkspace(ctx, &HostedAgentWorkspaceCreateRequest{SizeGibibytes: 20})
			require.Error(t, err)
			assert.Nil(t, got)
			var errResp *ErrorResponse
			require.ErrorAs(t, err, &errResp)
			assert.Equal(t, tt.status, errResp.Response.StatusCode)
			assert.Equal(t, tt.message, errResp.Message)
			require.NotNil(t, resp)
			assert.Equal(t, tt.status, resp.StatusCode)
		})
	}
}

func TestHostedAgents_GetWorkspace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/workspaces/"+hostedAgentWorkspaceTestID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "/v2/agents/workspaces/"+hostedAgentWorkspaceTestID, r.URL.Path)
		fmt.Fprintf(w, `{"workspace":%s}`, hostedAgentWorkspaceJSON)
	})

	got, resp, err := client.HostedAgents.GetWorkspace(ctx, hostedAgentWorkspaceTestID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, hostedAgentWorkspace, *got)
	assert.Equal(t, HostedAgentWorkspaceStateAttached, got.State)
	assert.Equal(t, "sess-abc123", got.AttachedSessionID)
	assert.Equal(t, int32(20), got.SizeGibibytes)
	assert.Equal(t, int64(734003200), got.BytesUsed)
	require.NotNil(t, got.LastSavedAt)
}

func TestHostedAgents_GetWorkspace_OptionalFieldsAbsent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/workspaces/"+hostedAgentWorkspaceTestID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"workspace":{
			"workspace_id": "019fb39c-14d9-7080-933e-b9b90e25acda",
			"state": "AVAILABLE",
			"size_gibibytes": 5,
			"bytes_used": 0,
			"created_at": "2026-10-01T08:00:00Z",
			"updated_at": "2026-10-01T08:00:00Z"
		}}`)
	})

	got, _, err := client.HostedAgents.GetWorkspace(ctx, hostedAgentWorkspaceTestID)
	require.NoError(t, err)
	assert.Equal(t, HostedAgentWorkspaceStateAvailable, got.State)
	assert.Empty(t, got.Name)
	assert.Empty(t, got.AttachedSessionID)
	assert.Nil(t, got.LastSavedAt)
	assert.Equal(t, int32(5), got.SizeGibibytes)
	assert.Equal(t, int64(0), got.BytesUsed)
}

func TestHostedAgentWorkspace_DecodesAllStates(t *testing.T) {
	tests := []struct {
		wire string
		want HostedAgentWorkspaceState
	}{
		{"AVAILABLE", HostedAgentWorkspaceStateAvailable},
		{"ATTACHING", HostedAgentWorkspaceStateAttaching},
		{"ATTACHED", HostedAgentWorkspaceStateAttached},
		{"RELEASING", HostedAgentWorkspaceStateReleasing},
		{"FAILED", HostedAgentWorkspaceStateFailed},
	}
	for _, tt := range tests {
		t.Run(tt.wire, func(t *testing.T) {
			var ws HostedAgentWorkspace
			require.NoError(t, json.Unmarshal([]byte(fmt.Sprintf(`{"state":%q}`, tt.wire)), &ws))
			assert.Equal(t, tt.want, ws.State)
			assert.Equal(t, tt.wire, string(tt.want))
		})
	}
}

func TestHostedAgents_GetWorkspace_NoWorkspaceInResponse(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/workspaces/"+hostedAgentWorkspaceTestID, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	})

	got, _, err := client.HostedAgents.GetWorkspace(ctx, hostedAgentWorkspaceTestID)
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestHostedAgents_GetWorkspace_EmptyID(t *testing.T) {
	got, resp, err := client.HostedAgents.GetWorkspace(ctx, "")
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Nil(t, resp)
}

func TestHostedAgents_GetWorkspace_NotFound(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/workspaces/"+hostedAgentWorkspaceTestID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"id":"not_found","message":"workspace not found"}`)
	})

	got, _, err := client.HostedAgents.GetWorkspace(ctx, hostedAgentWorkspaceTestID)
	require.Error(t, err)
	assert.Nil(t, got)
	var errResp *ErrorResponse
	require.ErrorAs(t, err, &errResp)
	assert.Equal(t, http.StatusNotFound, errResp.Response.StatusCode)
}

func TestHostedAgents_ListWorkspaces(t *testing.T) {
	tests := []struct {
		name     string
		opt      *HostedAgentWorkspaceListOptions
		wantSize string
		wantTok  string
	}{
		{name: "nil options"},
		{name: "empty options", opt: &HostedAgentWorkspaceListOptions{}},
		{name: "page size only", opt: &HostedAgentWorkspaceListOptions{PageSize: 25}, wantSize: "25"},
		{name: "page token only", opt: &HostedAgentWorkspaceListOptions{PageToken: "tok-1"}, wantTok: "tok-1"},
		{name: "both", opt: &HostedAgentWorkspaceListOptions{PageSize: 25, PageToken: "tok-1"}, wantSize: "25", wantTok: "tok-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc("/v2/agents/workspaces", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				q := r.URL.Query()
				assert.Equal(t, tt.wantSize != "", q.Has("page_size"))
				assert.Equal(t, tt.wantTok != "", q.Has("page_token"))
				assert.Equal(t, tt.wantSize, q.Get("page_size"))
				assert.Equal(t, tt.wantTok, q.Get("page_token"))
				fmt.Fprintf(w, `{"workspaces":[%s],"next_page_token":"tok-2"}`, hostedAgentWorkspaceJSON)
			})

			got, resp, err := client.HostedAgents.ListWorkspaces(ctx, tt.opt)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			require.Len(t, got.Workspaces, 1)
			assert.Equal(t, hostedAgentWorkspace, got.Workspaces[0])
			assert.Equal(t, "tok-2", got.NextPageToken)
		})
	}
}

func TestHostedAgents_ListWorkspaces_Empty(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/workspaces", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"workspaces":[],"next_page_token":""}`)
	})

	got, _, err := client.HostedAgents.ListWorkspaces(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Empty(t, got.Workspaces)
	assert.Empty(t, got.NextPageToken)
}

func TestHostedAgents_DeleteWorkspace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/workspaces/"+hostedAgentWorkspaceTestID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		assert.Equal(t, "/v2/agents/workspaces/"+hostedAgentWorkspaceTestID, r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.DeleteWorkspace(ctx, hostedAgentWorkspaceTestID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHostedAgents_DeleteWorkspace_EmptyID(t *testing.T) {
	resp, err := client.HostedAgents.DeleteWorkspace(ctx, "")
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestHostedAgents_DeleteWorkspace_APIErrors(t *testing.T) {
	tests := []struct {
		status  int
		message string
	}{
		{status: http.StatusNotFound, message: "workspace not found"},
		{status: http.StatusConflict, message: "workspace is in use by a session"},
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc("/v2/agents/workspaces/"+hostedAgentWorkspaceTestID, func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodDelete)
				w.WriteHeader(tt.status)
				fmt.Fprintf(w, `{"id":"err","message":%q}`, tt.message)
			})

			resp, err := client.HostedAgents.DeleteWorkspace(ctx, hostedAgentWorkspaceTestID)
			require.Error(t, err)
			var errResp *ErrorResponse
			require.ErrorAs(t, err, &errResp)
			assert.Equal(t, tt.status, errResp.Response.StatusCode)
			assert.Equal(t, tt.message, errResp.Message)
			require.NotNil(t, resp)
			assert.Equal(t, tt.status, resp.StatusCode)
		})
	}
}

func TestHostedAgentSession_DecodesWorkspaceID(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{
			name: "present",
			json: `{"session_id":"sess-ws","workspace_id":"019fb39c-14d9-7080-933e-b9b90e25acda"}`,
			want: hostedAgentWorkspaceTestID,
		},
		{
			name: "absent",
			json: `{"session_id":"sess-no-ws"}`,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var session HostedAgentSession
			require.NoError(t, json.Unmarshal([]byte(tt.json), &session))
			assert.Equal(t, tt.want, session.WorkspaceID)

			body, err := json.Marshal(&session)
			require.NoError(t, err)
			var fields map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(body, &fields))
			if tt.want == "" {
				assert.NotContains(t, fields, "workspace_id")
			} else {
				assert.Contains(t, fields, "workspace_id")
			}
		})
	}
}

func TestHostedAgents_GetSession_WorkspaceID(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-ws", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"session":{"session_id":"sess-ws","workspace_id":"019fb39c-14d9-7080-933e-b9b90e25acda"}}`)
	})

	got, _, err := client.HostedAgents.GetSession(ctx, "sess-ws")
	require.NoError(t, err)
	assert.Equal(t, hostedAgentWorkspaceTestID, got.WorkspaceID)
}

func TestHostedAgents_CreateSessionFromConfig_WorkspaceID(t *testing.T) {
	tests := []struct {
		name        string
		workspaceID string
		wantInBody  bool
	}{
		{name: "set", workspaceID: hostedAgentWorkspaceTestID, wantInBody: true},
		{name: "empty", workspaceID: "", wantInBody: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodPost)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.False(t, r.URL.Query().Has("workspace_id"), "config create sends the workspace in the JSON body, not as a query param")

				raw, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				var fields map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(raw, &fields))
				if tt.wantInBody {
					assert.JSONEq(t, fmt.Sprintf("%q", tt.workspaceID), string(fields["workspace_id"]))
				} else {
					assert.NotContains(t, fields, "workspace_id")
				}

				fmt.Fprintf(w, `{"session":{"session_id":"sess-cfg-ws","workspace_id":%q}}`, tt.workspaceID)
			})

			session, resp, err := client.HostedAgents.CreateSessionFromConfig(ctx, &HostedAgentSessionFromConfigRequest{
				Name:        "session-from-config",
				ConfigID:    "019fb39c-14d9-7080-933e-b9b90e25acda",
				WorkspaceID: tt.workspaceID,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.workspaceID, session.WorkspaceID)
		})
	}
}

func TestHostedAgents_CreateSessionFromManifest_NoWorkspaceIDQuery(t *testing.T) {
	const manifest = `name: probe
agent: opencode
`
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "application/x-yaml", r.Header.Get("Content-Type"))
		// A workspace cannot be attached from a manifest, so no workspace_id is sent.
		assert.False(t, r.URL.Query().Has("workspace_id"))

		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, manifest, string(raw))

		fmt.Fprint(w, `{"session":{"session_id":"sess-man"}}`)
	})

	got, resp, err := client.HostedAgents.CreateSessionFromManifest(ctx, []byte(manifest), &HostedAgentManifestCreateOptions{
		ResumeOnTopoff:  true,
		OpenAISessionID: "sess_abc",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "sess-man", got.SessionID)
	assert.Empty(t, got.WorkspaceID)
}

func TestHostedAgentSessionFromConfigRequest_JSONIncludesWorkspaceID(t *testing.T) {
	body, err := json.Marshal(&HostedAgentSessionFromConfigRequest{
		Name:        "session-from-config",
		ConfigID:    "019fb39c-14d9-7080-933e-b9b90e25acda",
		WorkspaceID: hostedAgentWorkspaceTestID,
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"session-from-config","config_id":"019fb39c-14d9-7080-933e-b9b90e25acda","workspace_id":"019fb39c-14d9-7080-933e-b9b90e25acda"}`, string(body))
}
