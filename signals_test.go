package godo

import (
	"encoding/json"
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
