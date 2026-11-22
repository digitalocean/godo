package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

const sandboxTemplateJSON = `{"template_id":"t-1","name":"my-template","status":"TEMPLATE_STATUS_READY","spec":{"base_template":"coding-base","image":{"registry":"registry.digitalocean.com","repository":"my-team/my-image","tag":"v1"}},"created_at":"2026-11-20T12:00:00Z","updated_at":"2026-11-20T12:05:00Z"}`

func sandboxTemplateWant(status SandboxTemplateStatus) *SandboxTemplate {
	return &SandboxTemplate{
		ID:     "t-1",
		Name:   "my-template",
		Status: status,
		Spec: &SandboxTemplateSpec{
			BaseTemplate: "coding-base",
			Image:        &SandboxImageSource{Registry: "registry.digitalocean.com", Repository: "my-team/my-image", Tag: "v1"},
		},
		CreatedAt: Timestamp{time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)},
		UpdatedAt: Timestamp{time.Date(2026, 11, 20, 12, 5, 0, 0, time.UTC)},
	}
}

func TestSandboxes_CreateTemplate(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/templates", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		want := map[string]interface{}{
			"name":           "my-template",
			"base_template":  "coding-base",
			"source_oci_ref": "registry.digitalocean.com/my-team/my-image:v1",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("request body = %+v, expected %+v", got, want)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"template":%s}`, sandboxTemplateJSON)
	})

	tpl, _, err := client.Sandboxes.CreateTemplate(ctx, &SandboxTemplateCreateRequest{
		Name: "my-template", BaseTemplate: "coding-base", SourceOCIRef: "registry.digitalocean.com/my-team/my-image:v1",
	})
	if err != nil || !reflect.DeepEqual(tpl, sandboxTemplateWant(SandboxTemplateStatusReady)) {
		t.Fatalf("CreateTemplate returned %+v, %v", tpl, err)
	}
	if _, _, err := client.Sandboxes.CreateTemplate(ctx, nil); err == nil {
		t.Error("a nil request returned no error")
	}
}

func TestSandboxes_GetUpdateDeleteTemplate(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/templates/t-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(w, `{"template":%s}`, sandboxTemplateJSON)
		case http.MethodPut:
			var got map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&got)
			if !reflect.DeepEqual(got, map[string]interface{}{"source_oci_ref": "r/i:v2"}) {
				t.Errorf("update body = %+v, expected only source_oci_ref", got)
			}
			fmt.Fprintf(w, `{"template":%s}`, sandboxTemplateJSON)
		case http.MethodDelete:
			fmt.Fprint(w, `{}`)
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})

	got, _, err := client.Sandboxes.GetTemplate(ctx, "t-1")
	if err != nil || !reflect.DeepEqual(got, sandboxTemplateWant(SandboxTemplateStatusReady)) {
		t.Fatalf("GetTemplate returned %+v, %v", got, err)
	}
	if _, _, err := client.Sandboxes.UpdateTemplate(ctx, "t-1", &SandboxTemplateUpdateRequest{SourceOCIRef: "r/i:v2"}); err != nil {
		t.Fatalf("UpdateTemplate returned error: %v", err)
	}
	if _, err := client.Sandboxes.DeleteTemplate(ctx, "t-1"); err != nil {
		t.Fatalf("DeleteTemplate returned error: %v", err)
	}
	for name, err := range map[string]error{
		"get": func() error { _, _, e := client.Sandboxes.GetTemplate(ctx, ""); return e }(),
		"update": func() error {
			_, _, e := client.Sandboxes.UpdateTemplate(ctx, "", &SandboxTemplateUpdateRequest{})
			return e
		}(),
		"update nil": func() error { _, _, e := client.Sandboxes.UpdateTemplate(ctx, "t-1", nil); return e }(),
		"delete":     func() error { _, e := client.Sandboxes.DeleteTemplate(ctx, ""); return e }(),
	} {
		if err == nil {
			t.Errorf("%s: no error for missing input", name)
		}
	}
}

func TestSandboxes_ListTemplates(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/templates", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{"page_size": "1", "page_token": "t1"})
		fmt.Fprintf(w, `{"templates":[%s],"next_page_token":"t2"}`, sandboxTemplateJSON)
	})
	list, _, err := client.Sandboxes.ListTemplates(ctx, &SandboxTemplateListOptions{PageSize: 1, PageToken: "t1"})
	want := &SandboxTemplateList{Templates: []SandboxTemplate{*sandboxTemplateWant(SandboxTemplateStatusReady)}, NextPageToken: "t2"}
	if err != nil || !reflect.DeepEqual(list, want) {
		t.Fatalf("ListTemplates returned %+v, %v; expected %+v", list, err, want)
	}
}

func TestSandboxes_TemplateBuilds(t *testing.T) {
	setup()
	defer teardown()

	const build = `{"build_id":"b-1","template_id":"t-1","name":"my-template","status":"TEMPLATE_BUILD_STATUS_FAILED","error":"image not found","created_at":"2026-11-20T12:00:00Z","updated_at":"2026-11-20T12:01:00Z"}`
	mux.HandleFunc("/v2/sandboxes/templates/t-1/builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{"page_size": "5"})
		fmt.Fprintf(w, `{"builds":[%s],"next_page_token":""}`, build)
	})
	mux.HandleFunc("/v2/sandboxes/templates/t-1/builds/b-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"build":%s}`, build)
	})

	list, _, err := client.Sandboxes.ListTemplateBuilds(ctx, "t-1", &SandboxTemplateListOptions{PageSize: 5})
	if err != nil || len(list.Builds) != 1 || list.Builds[0].ID != "b-1" || list.Builds[0].Status != SandboxTemplateBuildStatusFailed ||
		list.Builds[0].Error != "image not found" || list.NextPageToken != "" {
		t.Fatalf("ListTemplateBuilds returned %+v, %v", list, err)
	}
	got, _, err := client.Sandboxes.GetTemplateBuild(ctx, "t-1", "b-1")
	if err != nil || got.TemplateID != "t-1" || got.Status != SandboxTemplateBuildStatusFailed {
		t.Fatalf("GetTemplateBuild returned %+v, %v", got, err)
	}
	if _, _, err := client.Sandboxes.GetTemplateBuild(ctx, "t-1", ""); err == nil {
		t.Error("an empty build id returned no error")
	}
	if _, _, err := client.Sandboxes.ListTemplateBuilds(ctx, "", nil); err == nil {
		t.Error("an empty template id returned no error")
	}
}

func TestSandboxes_TemplateErrors(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/templates", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"id":"conflict","message":"name collides with a base template","request_id":"req-1"}`)
	})
	_, _, err := client.Sandboxes.CreateTemplate(ctx, &SandboxTemplateCreateRequest{Name: "base", BaseTemplate: "b", SourceOCIRef: "r"})
	apiErr, ok := err.(*ErrorResponse)
	if !ok || apiErr.Response.StatusCode != http.StatusConflict || apiErr.RequestID != "req-1" {
		t.Fatalf("CreateTemplate returned %#v, expected a 409 ErrorResponse", err)
	}
}
