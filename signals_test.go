package godo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestSignals_GetAgentConsent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/consent/agent-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"team_id":123,"agent_id":"agent-1","enabled":true,"allowed":true}`)
	})

	got, _, err := client.Signals.GetAgentConsent(ctx, "agent-1")
	if err != nil {
		t.Fatalf("GetAgentConsent: %v", err)
	}
	want := &SignalsAgentConsent{TeamID: 123, AgentID: "agent-1", Enabled: true, Allowed: true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v want %+v", got, want)
	}
}

func TestSignals_SetAgentConsent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/consent/agent-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var req signalsSetConsentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !req.Enabled {
			t.Errorf("enabled=%v", req.Enabled)
		}
		fmt.Fprint(w, `{"consent":{"id":1,"team_id":123,"agent_id":"agent-1","enabled":true,"updated_at":"2026-10-05T12:00:00Z"}}`)
	})

	got, _, err := client.Signals.SetAgentConsent(ctx, "agent-1", true)
	if err != nil {
		t.Fatalf("SetAgentConsent: %v", err)
	}
	if got == nil || got.AgentID != "agent-1" || !got.Enabled || got.ID != 1 {
		t.Errorf("unexpected consent: %+v", got)
	}
}

func TestSignals_ListAgentSessions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/agents/agent-1/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"limit":       "20",
			"after":       "c1",
			"start_time":  "1727740800",
			"signal_type": "Looping",
		})
		fmt.Fprint(w, `{
			"edges":[{"cursor":"c1","node":{"session_id":"sess-1","total_turns":12,"started_at":"2026-10-01T12:00:00Z","duration_seconds":340,"signal_count":3}}],
			"page_info":{"has_next_page":true,"end_cursor":"c1"}
		}`)
	})

	start := int64(1727740800)
	got, _, err := client.Signals.ListAgentSessions(ctx, "agent-1", &SignalsListAgentSessionsOptions{
		SignalsCursorPageOptions: SignalsCursorPageOptions{Limit: 20, After: "c1"},
		StartTime:                &start,
		SignalType:               []string{"Looping"},
	})
	if err != nil {
		t.Fatalf("ListAgentSessions: %v", err)
	}
	if got.PageInfo.EndCursor != "c1" || !got.PageInfo.HasNextPage {
		t.Errorf("unexpected page_info: %+v", got.PageInfo)
	}
	if got.Edges[0].Node.SessionID != "sess-1" || got.Edges[0].Node.SignalCount != 3 {
		t.Errorf("unexpected node: %+v", got.Edges[0].Node)
	}
}

func TestSignals_ListSessionSegments(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/sessions/sess-1/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"edges":[{"cursor":"c","node":{"segment_id":"seg-1","session_id":"sess-1","segment_seq":1,"started_at":"2026-10-01T12:00:00Z","status":"closed","annotation_status":"done","total_turns":8,"duration_seconds":300,"signals":[{"signal_type":"Looping"}]}}],
			"page_info":{"has_next_page":false}
		}`)
	})

	got, _, err := client.Signals.ListSessionSegments(ctx, "sess-1", nil)
	if err != nil {
		t.Fatalf("ListSessionSegments: %v", err)
	}
	if got.Edges[0].Node.SegmentID != "seg-1" {
		t.Errorf("unexpected segment: %+v", got.Edges[0].Node)
	}
}

func TestSignals_ListSessionDialogues(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/sessions/sess-1/dialogues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"session_id":"sess-1",
			"edges":[{"cursor":"c","node":{
				"id":901,"run_id":"run-1","created_at":"2026-10-01T12:00:01Z","sequence":1,
				"segment_id":"seg-1","segment_seq":1,"user_message":"resize",
				"steps":[{"type":"tool_call"}],"run_status":"completed",
				"signals":[{"signal_type":"Looping","label":"loop","run_uuid":"run-1","step_index":0,"confidence":0.9,"layer":"observation","category":"execution"}]
			}}],
			"page_info":{"has_next_page":false}
		}`)
	})

	got, _, err := client.Signals.ListSessionDialogues(ctx, "sess-1", nil)
	if err != nil {
		t.Fatalf("ListSessionDialogues: %v", err)
	}
	if got.SessionID != "sess-1" || got.Edges[0].Node.SegmentID != "seg-1" {
		t.Errorf("unexpected dialogue: %+v", got.Edges[0].Node)
	}
	if len(got.Edges[0].Node.Steps) != 1 {
		t.Errorf("expected one step, got %d", len(got.Edges[0].Node.Steps))
	}
	if got.Edges[0].Node.Signals[0].SignalType != "Looping" {
		t.Errorf("unexpected signal: %+v", got.Edges[0].Node.Signals[0])
	}
}

func TestSignals_GetSegment(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/segments/seg-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"segment_id":"seg-1","session_id":"sess-1",
			"edges":[{"cursor":"c","node":{"id":901,"run_id":"run-1","created_at":"2026-10-01T12:00:01Z","sequence":1,"user_message":"resize","steps":[],"run_status":"completed"}}],
			"page_info":{"has_next_page":false}
		}`)
	})

	got, _, err := client.Signals.GetSegment(ctx, "seg-1", &SignalsCursorPageOptions{Limit: 20})
	if err != nil {
		t.Fatalf("GetSegment: %v", err)
	}
	if got.SegmentID != "seg-1" || got.Edges[0].Node.ID != 901 {
		t.Errorf("unexpected detail: %+v", got)
	}
}

func TestSignals_GetSegmentSignalReport(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/segments/seg-1/signal-report", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"segment_id":"seg-1","session_id":"sess-1","quality_score":0.82,"concerning":false,
			"total_turns":8,"user_turns":4,"assistant_turns":4,"is_dragging":false,"groups":[]
		}`)
	})

	got, _, err := client.Signals.GetSegmentSignalReport(ctx, "seg-1")
	if err != nil {
		t.Fatalf("GetSegmentSignalReport: %v", err)
	}
	if got.QualityScore != 0.82 || got.TotalTurns != 8 {
		t.Errorf("unexpected report: %+v", got)
	}
}

func TestSignals_Exports(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/exports/options", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"filters":{"signal_type":["Looping","Hallucination"]}}`)
	})
	mux.HandleFunc("/v1/signals/exports/exp-1/download", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"download_url":"https://presigned","expires_at":1728000900}`)
	})
	mux.HandleFunc("/v1/signals/exports/exp-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"export_id":"exp-1","agent_id":"agent-1","status":"complete","created_at":1728000000,"filters":{}}`)
	})
	mux.HandleFunc("/v1/signals/exports", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("agent_id"); got != "agent-1" {
			t.Errorf("agent_id=%s", got)
		}
		fmt.Fprint(w, `{"edges":[{"cursor":"c","node":{"export_id":"exp-1","status":"complete","created_at":1,"filters":{}}}],"page_info":{"has_next_page":false}}`)
	})

	opts, _, err := client.Signals.GetExportOptions(ctx)
	if err != nil {
		t.Fatalf("GetExportOptions: %v", err)
	}
	if !reflect.DeepEqual(opts.Filters.SignalType, []string{"Looping", "Hallucination"}) {
		t.Errorf("unexpected options: %+v", opts)
	}

	list, _, err := client.Signals.ListExports(ctx, &SignalsListExportsOptions{AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("ListExports: %v", err)
	}
	if list.Edges[0].Node.ExportID != "exp-1" {
		t.Errorf("unexpected list: %+v", list)
	}

	job, _, err := client.Signals.GetExport(ctx, "exp-1")
	if err != nil {
		t.Fatalf("GetExport: %v", err)
	}
	if job.Status != "complete" {
		t.Errorf("unexpected job: %+v", job)
	}

	dl, _, err := client.Signals.GetExportDownload(ctx, "exp-1")
	if err != nil {
		t.Fatalf("GetExportDownload: %v", err)
	}
	if dl.DownloadURL != "https://presigned" || dl.ExpiresAt != 1728000900 {
		t.Errorf("unexpected download: %+v", dl)
	}
}

func TestSignals_DialogueStepsRawJSON(t *testing.T) {
	raw := json.RawMessage(`{"type":"tool_call"}`)
	d := SignalsDialogue{Steps: []json.RawMessage{raw}}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(b) {
		t.Fatalf("invalid json: %s", b)
	}
}

func TestSignals_ListConsents(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/consent", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"team_id":123,
			"consents":[
				{"id":1,"team_id":123,"agent_id":"agent-1","enabled":true,"updated_at":"2026-10-05T12:00:00Z"},
				{"id":2,"team_id":123,"agent_id":"agent-2","enabled":false,"updated_at":"2026-10-05T13:00:00Z"}
			]
		}`)
	})

	got, _, err := client.Signals.ListConsents(ctx)
	if err != nil {
		t.Fatalf("ListConsents: %v", err)
	}
	if got.TeamID != 123 {
		t.Errorf("team_id=%d, want 123", got.TeamID)
	}
	if len(got.Consents) != 2 {
		t.Fatalf("len(consents)=%d, want 2", len(got.Consents))
	}
	if got.Consents[0].AgentID != "agent-1" || !got.Consents[0].Enabled {
		t.Errorf("unexpected consent[0]: %+v", got.Consents[0])
	}
	if got.Consents[1].AgentID != "agent-2" || got.Consents[1].Enabled {
		t.Errorf("unexpected consent[1]: %+v", got.Consents[1])
	}
}

func TestSignals_ListConsents_Empty(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/consent", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"team_id":123,"consents":[]}`)
	})

	got, _, err := client.Signals.ListConsents(ctx)
	if err != nil {
		t.Fatalf("ListConsents: %v", err)
	}
	if len(got.Consents) != 0 {
		t.Errorf("expected empty consents, got %d", len(got.Consents))
	}
}

func TestSignals_CreateExport(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/exports", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body SignalsCreateExportRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.AgentID != "agent-1" {
			t.Errorf("agent_id=%s, want agent-1", body.AgentID)
		}
		if !reflect.DeepEqual(body.SignalType, []string{"Looping"}) {
			t.Errorf("signal_type=%v, want [Looping]", body.SignalType)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"export_id":"exp-new","agent_id":"agent-1","status":"pending","created_at":1728000000,"filters":{"signal_type":["Looping"]}}`)
	})

	got, resp, err := client.Signals.CreateExport(ctx, &SignalsCreateExportRequest{
		AgentID:    "agent-1",
		SignalType: []string{"Looping"},
	})
	if err != nil {
		t.Fatalf("CreateExport: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status=%d, want 201", resp.StatusCode)
	}
	if got.ExportID != "exp-new" || got.Status != "pending" {
		t.Errorf("unexpected export: %+v", got)
	}
}

func TestSignals_CreateExport_WithTimeRange(t *testing.T) {
	setup()
	defer teardown()

	start := int64(1727740800)
	end := int64(1728000000)

	mux.HandleFunc("/v1/signals/exports", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body SignalsCreateExportRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.AgentID != "agent-1" {
			t.Errorf("agent_id=%s", body.AgentID)
		}
		if body.StartTime == nil || *body.StartTime != start {
			t.Errorf("start_time=%v, want %d", body.StartTime, start)
		}
		if body.EndTime == nil || *body.EndTime != end {
			t.Errorf("end_time=%v, want %d", body.EndTime, end)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"export_id":"exp-time","agent_id":"agent-1","status":"pending","created_at":1728000000,"filters":{"start_time":1727740800,"end_time":1728000000}}`)
	})

	got, _, err := client.Signals.CreateExport(ctx, &SignalsCreateExportRequest{
		AgentID:   "agent-1",
		StartTime: &start,
		EndTime:   &end,
	})
	if err != nil {
		t.Fatalf("CreateExport: %v", err)
	}
	if got.ExportID != "exp-time" {
		t.Errorf("unexpected export: %+v", got)
	}
}

func TestSignals_CreateExport_NilRequest(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Signals.CreateExport(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestSignals_CreateExport_EmptyAgentID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Signals.CreateExport(ctx, &SignalsCreateExportRequest{})
	if err == nil {
		t.Fatal("expected error for empty agent_id")
	}
}

const testDeletionID = "01JABCDEFGHJKMNPQRSTVWXYZ0"

// registerNoCallHandler fails the test if the path is hit and returns a counter.
func registerNoCallHandler(t *testing.T, pattern string) *int {
	t.Helper()
	calls := 0
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		calls++
		t.Errorf("unexpected HTTP call: %s %s", r.Method, r.URL.String())
		w.WriteHeader(http.StatusInternalServerError)
	})
	return &calls
}

func assertErrStatus(t *testing.T, err error, resp *Response, want int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with status %d", want)
	}
	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T: %v", err, err)
	}
	if errResp.Response.StatusCode != want {
		t.Errorf("error status=%d, want %d", errResp.Response.StatusCode, want)
	}
	if resp == nil || resp.StatusCode != want {
		t.Errorf("resp=%v, want status %d", resp, want)
	}
}

func TestSignals_CreateDeletion_ManagedAgent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body SignalsCreateDeletionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		want := SignalsCreateDeletionRequest{Type: "managed_agent", TeamID: 12345, AgentID: "agt-uuid"}
		if body != want {
			t.Errorf("body=%+v, want %+v", body, want)
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"team_id":12345,"deletion_id":"`+testDeletionID+`","type":"managed_agent","agent_id":"agt-uuid","status":"queued","error_message":null,"created_at":1728324000,"started_at":null,"completed_at":null}`)
	})

	got, resp, err := client.Signals.CreateDeletion(ctx, &SignalsCreateDeletionRequest{
		Type: SignalsDeletionTypeManagedAgent, TeamID: 12345, AgentID: "agt-uuid",
	})
	if err != nil {
		t.Fatalf("CreateDeletion: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("status=%d, want 202", resp.StatusCode)
	}
	want := &SignalsDeletionJob{
		TeamID: 12345, DeletionID: testDeletionID, Type: "managed_agent", AgentID: "agt-uuid",
		Status: SignalsDeletionStatusQueued, CreatedAt: 1728324000,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v want %+v", got, want)
	}
}

func TestSignals_CreateDeletion_Inference(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var m map[string]any
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		if m["type"] != "inference" {
			t.Errorf("type=%v, want inference", m["type"])
		}
		if m["team_id"] != float64(12345) {
			t.Errorf("team_id=%v, want 12345", m["team_id"])
		}
		if _, has := m["agent_id"]; has {
			t.Errorf("agent_id must not be sent for inference: %v", m)
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"team_id":12345,"deletion_id":"`+testDeletionID+`","type":"inference","status":"queued","created_at":1728324000}`)
	})

	got, resp, err := client.Signals.CreateDeletion(ctx, &SignalsCreateDeletionRequest{
		Type: SignalsDeletionTypeInference, TeamID: 12345,
	})
	if err != nil {
		t.Fatalf("CreateDeletion: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("status=%d, want 202", resp.StatusCode)
	}
	if got.Type != "inference" || got.AgentID != "" {
		t.Errorf("unexpected job: %+v", got)
	}
}

func TestSignals_CreateDeletion_ExistingActiveJob(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"team_id":12345,"deletion_id":"`+testDeletionID+`","type":"managed_agent","agent_id":"agt-uuid","status":"running","created_at":1728324000,"started_at":1728324010}`)
	})

	got, resp, err := client.Signals.CreateDeletion(ctx, &SignalsCreateDeletionRequest{
		Type: SignalsDeletionTypeManagedAgent, TeamID: 12345, AgentID: "agt-uuid",
	})
	if err != nil {
		t.Fatalf("CreateDeletion: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status=%d, want 200", resp.StatusCode)
	}
	if got.DeletionID != testDeletionID || got.Status != SignalsDeletionStatusRunning {
		t.Errorf("unexpected job: %+v", got)
	}
	if got.StartedAt == nil || *got.StartedAt != 1728324010 {
		t.Errorf("started_at=%v", got.StartedAt)
	}
}

func TestSignals_CreateDeletion_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *SignalsCreateDeletionRequest
		wantMsg string
	}{
		{"NilRequest", nil, "signals: create deletion request is required"},
		{"ZeroTeamID", &SignalsCreateDeletionRequest{Type: "inference", TeamID: 0}, "signals: team_id is required"},
		{"NegativeTeamID", &SignalsCreateDeletionRequest{Type: "inference", TeamID: -1}, "signals: team_id is required"},
		{"EmptyType", &SignalsCreateDeletionRequest{TeamID: 1}, "signals: type must be managed_agent or inference"},
		{"UnknownType", &SignalsCreateDeletionRequest{Type: "agent", TeamID: 1, AgentID: "a"}, "signals: type must be managed_agent or inference"},
		{"BadTypeAndBadTeamReportsTeamFirst", &SignalsCreateDeletionRequest{Type: "agent", TeamID: 0}, "signals: team_id is required"},
		{"ManagedAgentMissingAgentID", &SignalsCreateDeletionRequest{Type: "managed_agent", TeamID: 1}, "signals: agent_id is required for managed_agent"},
		{"InferenceWithAgentID", &SignalsCreateDeletionRequest{Type: "inference", TeamID: 1, AgentID: "a"}, "signals: agent_id must not be set for inference"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			defer teardown()
			calls := registerNoCallHandler(t, "/v1/signals/deletions")

			got, resp, err := client.Signals.CreateDeletion(ctx, tt.req)
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != tt.wantMsg {
				t.Errorf("error=%q, want %q", err.Error(), tt.wantMsg)
			}
			if got != nil || resp != nil {
				t.Errorf("expected nil job and response, got %+v / %+v", got, resp)
			}
			if *calls != 0 {
				t.Errorf("HTTP calls=%d, want 0", *calls)
			}
		})
	}
}

func TestSignals_CreateDeletion_DoesNotMutateRequest(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"deletion_id":"x","status":"queued"}`)
	})
	req := &SignalsCreateDeletionRequest{Type: "managed_agent", TeamID: 7, AgentID: " agt "}
	orig := *req
	if _, _, err := client.Signals.CreateDeletion(ctx, req); err != nil {
		t.Fatalf("CreateDeletion: %v", err)
	}
	if *req != orig {
		t.Errorf("request mutated: %+v", *req)
	}
}

func TestSignals_CreateDeletion_ErrorStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"Forbidden", http.StatusForbidden, `{"id":"forbidden","message":"team mismatch"}`},
		{"AgentNotFound", http.StatusNotFound, `{"id":"not_found","message":"agent not found"}`},
		{"TooManyActive", http.StatusTooManyRequests, `{"id":"too_many_requests","message":"too many active deletions"}`},
		{"ServerError", http.StatusInternalServerError, `{"id":"server_error","message":"boom"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			defer teardown()
			mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodPost)
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			})

			got, resp, err := client.Signals.CreateDeletion(ctx, &SignalsCreateDeletionRequest{
				Type: SignalsDeletionTypeManagedAgent, TeamID: 1, AgentID: "agt",
			})
			if got != nil {
				t.Errorf("expected nil job, got %+v", got)
			}
			assertErrStatus(t, err, resp, tt.status)
		})
	}
}

func TestSignals_CreateDeletion_ForbiddenMessage(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"id":"forbidden","message":"team mismatch"}`)
	})
	_, _, err := client.Signals.CreateDeletion(ctx, &SignalsCreateDeletionRequest{Type: "inference", TeamID: 1})
	if err == nil || !errors.As(err, new(*ErrorResponse)) {
		t.Fatalf("expected ErrorResponse, got %v", err)
	}
	var errResp *ErrorResponse
	errors.As(err, &errResp)
	if errResp.Message != "team mismatch" {
		t.Errorf("message=%q, want %q", errResp.Message, "team mismatch")
	}
}

func TestSignals_CreateDeletion_ContextCancelled(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{}`)
	})
	cctx, cancel := context.WithCancel(ctx)
	cancel()

	got, _, err := client.Signals.CreateDeletion(cctx, &SignalsCreateDeletionRequest{Type: "inference", TeamID: 1})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error=%v, want context.Canceled", err)
	}
	if got != nil {
		t.Errorf("expected nil job, got %+v", got)
	}
}

func TestSignals_CreateDeletion_EmptyBody(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	got, resp, err := client.Signals.CreateDeletion(ctx, &SignalsCreateDeletionRequest{Type: "inference", TeamID: 1})
	if err == nil {
		t.Fatal("expected decode error for empty body")
	}
	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}
	if got != nil {
		t.Errorf("expected nil job, got %+v", got)
	}
}

func TestSignals_GetDeletion_Queued(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions/"+testDeletionID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"team_id":12345,"deletion_id":"`+testDeletionID+`","type":"managed_agent","agent_id":"agt-uuid","status":"queued","error_message":null,"created_at":1728324000,"started_at":null,"completed_at":null,"extra_future_field":true}`)
	})

	got, resp, err := client.Signals.GetDeletion(ctx, testDeletionID)
	if err != nil {
		t.Fatalf("GetDeletion: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status=%d, want 200", resp.StatusCode)
	}
	if got.ErrorMessage != nil || got.StartedAt != nil || got.CompletedAt != nil {
		t.Errorf("expected nil optional fields: %+v", got)
	}
	if got.Status != SignalsDeletionStatusQueued || got.AgentID != "agt-uuid" {
		t.Errorf("unexpected job: %+v", got)
	}
}

func TestSignals_GetDeletion_Failed(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions/"+testDeletionID, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"team_id":1,"deletion_id":"`+testDeletionID+`","type":"inference","status":"failed","error_message":"max attempts exceeded","created_at":1,"started_at":2,"completed_at":3}`)
	})

	got, _, err := client.Signals.GetDeletion(ctx, testDeletionID)
	if err != nil {
		t.Fatalf("GetDeletion: %v", err)
	}
	if got.Status != SignalsDeletionStatusFailed {
		t.Errorf("status=%s", got.Status)
	}
	if got.ErrorMessage == nil || *got.ErrorMessage != "max attempts exceeded" {
		t.Errorf("error_message=%v", got.ErrorMessage)
	}
}

func TestSignals_GetDeletion_Complete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions/"+testDeletionID, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"team_id":1,"deletion_id":"`+testDeletionID+`","type":"managed_agent","agent_id":"a","status":"complete","created_at":10,"started_at":20,"completed_at":30}`)
	})

	got, _, err := client.Signals.GetDeletion(ctx, testDeletionID)
	if err != nil {
		t.Fatalf("GetDeletion: %v", err)
	}
	if got.Status != SignalsDeletionStatusComplete {
		t.Errorf("status=%s", got.Status)
	}
	if got.StartedAt == nil || *got.StartedAt != 20 || got.CompletedAt == nil || *got.CompletedAt != 30 || got.CreatedAt != 10 {
		t.Errorf("unexpected timestamps: %+v", got)
	}
}

func TestSignals_GetDeletion_UnknownStatus(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions/"+testDeletionID, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"deletion_id":"`+testDeletionID+`","status":"cancelled","created_at":1}`)
	})

	got, _, err := client.Signals.GetDeletion(ctx, testDeletionID)
	if err != nil {
		t.Fatalf("GetDeletion: %v", err)
	}
	if got.Status != "cancelled" {
		t.Errorf("status=%s, want cancelled", got.Status)
	}
}

func TestSignals_GetDeletion_EmptyID(t *testing.T) {
	for _, id := range []string{"", "   ", "\t\n"} {
		t.Run(fmt.Sprintf("%q", id), func(t *testing.T) {
			setup()
			defer teardown()
			listCalls := registerNoCallHandler(t, "/v1/signals/deletions")
			itemCalls := registerNoCallHandler(t, "/v1/signals/deletions/")

			got, resp, err := client.Signals.GetDeletion(ctx, id)
			if err == nil {
				t.Fatal("expected error")
			}
			if got != nil || resp != nil {
				t.Errorf("expected nil job and response")
			}
			if *listCalls != 0 || *itemCalls != 0 {
				t.Errorf("HTTP calls list=%d item=%d, want 0", *listCalls, *itemCalls)
			}
		})
	}
}

func TestSignals_GetDeletion_DotIDs(t *testing.T) {
	for _, id := range []string{".", "..", " . ", " .. "} {
		t.Run(fmt.Sprintf("%q", id), func(t *testing.T) {
			setup()
			defer teardown()
			listCalls := registerNoCallHandler(t, "/v1/signals/deletions")
			itemCalls := registerNoCallHandler(t, "/v1/signals/deletions/")

			_, resp, err := client.Signals.GetDeletion(ctx, id)
			if err == nil {
				t.Fatal("expected error")
			}
			if resp != nil {
				t.Errorf("expected nil response")
			}
			if *listCalls != 0 || *itemCalls != 0 {
				t.Errorf("HTTP calls list=%d item=%d, want 0", *listCalls, *itemCalls)
			}
		})
	}
}

func TestSignals_GetDeletion_EscapesID(t *testing.T) {
	tests := []struct {
		id          string
		wantEscaped string
	}{
		{"a/b", "/v1/signals/deletions/a%2Fb"},
		{"a?x=1", "/v1/signals/deletions/a%3Fx=1"},
		{" abc ", "/v1/signals/deletions/%20abc%20"},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			setup()
			defer teardown()

			called := false
			mux.HandleFunc("/v1/signals/deletions/", func(w http.ResponseWriter, r *http.Request) {
				called = true
				testMethod(t, r, http.MethodGet)
				if got := r.URL.EscapedPath(); got != tt.wantEscaped {
					t.Errorf("escaped path=%q, want %q", got, tt.wantEscaped)
				}
				if r.URL.RawQuery != "" {
					t.Errorf("unexpected query %q", r.URL.RawQuery)
				}
				fmt.Fprint(w, `{"deletion_id":"x","status":"queued","created_at":1}`)
			})

			if _, _, err := client.Signals.GetDeletion(ctx, tt.id); err != nil {
				t.Fatalf("GetDeletion: %v", err)
			}
			if !called {
				t.Error("handler not called")
			}
		})
	}
}

func TestSignals_GetDeletion_NotFound(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions/"+testDeletionID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"id":"not_found","message":"deletion not found"}`)
	})

	got, resp, err := client.Signals.GetDeletion(ctx, testDeletionID)
	if got != nil {
		t.Errorf("expected nil job, got %+v", got)
	}
	assertErrStatus(t, err, resp, http.StatusNotFound)
}

func TestSignals_ListDeletions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"edges":[
				{"cursor":"c1","node":{"team_id":1,"deletion_id":"`+testDeletionID+`","type":"managed_agent","agent_id":"a","status":"running","created_at":2}},
				{"cursor":"c2","node":{"team_id":1,"deletion_id":"01JABCDEFGHJKMNPQRSTVWXYZ1","type":"inference","status":"complete","created_at":1,"completed_at":5}}
			],
			"page_info":{"has_next_page":true,"end_cursor":"c2"}
		}`)
	})

	got, _, err := client.Signals.ListDeletions(ctx, nil)
	if err != nil {
		t.Fatalf("ListDeletions: %v", err)
	}
	if len(got.Edges) != 2 {
		t.Fatalf("edges=%d, want 2", len(got.Edges))
	}
	if !got.PageInfo.HasNextPage || got.PageInfo.EndCursor != "c2" {
		t.Errorf("unexpected page_info: %+v", got.PageInfo)
	}
	if got.Edges[0].Node.DeletionID != testDeletionID || got.Edges[1].Node.Type != "inference" || got.Edges[1].Node.AgentID != "" {
		t.Errorf("unexpected edges: %+v", got.Edges)
	}
}

func TestSignals_ListDeletions_NilOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if r.URL.RawQuery != "" {
			t.Errorf("unexpected query %q", r.URL.RawQuery)
		}
		fmt.Fprint(w, `{"edges":[],"page_info":{"has_next_page":false}}`)
	})

	if _, _, err := client.Signals.ListDeletions(ctx, nil); err != nil {
		t.Fatalf("ListDeletions: %v", err)
	}
}

func TestSignals_ListDeletions_WithOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{"limit": "50", "after": "abc"})
		fmt.Fprint(w, `{"edges":[],"page_info":{"has_next_page":false}}`)
	})

	_, _, err := client.Signals.ListDeletions(ctx, &SignalsListDeletionsOptions{
		SignalsCursorPageOptions: SignalsCursorPageOptions{Limit: 50, After: "abc"},
	})
	if err != nil {
		t.Fatalf("ListDeletions: %v", err)
	}
}

func TestSignals_ListDeletions_CursorRoundTrip(t *testing.T) {
	setup()
	defer teardown()

	const cursor = "YWJjZA=="
	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("after"); got != cursor {
			t.Errorf("after=%q, want %q", got, cursor)
		}
		fmt.Fprint(w, `{"edges":[],"page_info":{"has_next_page":false}}`)
	})

	_, _, err := client.Signals.ListDeletions(ctx, &SignalsListDeletionsOptions{
		SignalsCursorPageOptions: SignalsCursorPageOptions{After: cursor},
	})
	if err != nil {
		t.Fatalf("ListDeletions: %v", err)
	}
}

func TestSignals_ListDeletions_Empty(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"edges":[],"page_info":{"has_next_page":false}}`)
	})

	got, _, err := client.Signals.ListDeletions(ctx, nil)
	if err != nil {
		t.Fatalf("ListDeletions: %v", err)
	}
	if len(got.Edges) != 0 || got.PageInfo.HasNextPage {
		t.Errorf("unexpected response: %+v", got)
	}
}

func TestSignals_ListDeletions_BadCursor(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/signals/deletions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"id":"bad_request","message":"invalid cursor"}`)
	})

	got, resp, err := client.Signals.ListDeletions(ctx, &SignalsListDeletionsOptions{
		SignalsCursorPageOptions: SignalsCursorPageOptions{After: "garbage"},
	})
	if got != nil {
		t.Errorf("expected nil response body, got %+v", got)
	}
	assertErrStatus(t, err, resp, http.StatusBadRequest)
}
