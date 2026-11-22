package sandbox_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/sandbox"
)

func tplJSON(id string, status godo.SandboxTemplateStatus) string {
	return fmt.Sprintf(`{"template_id":%q,"name":"my-template","status":%q,"spec":{"base_template":"coding-base"},"created_at":"2026-11-20T12:00:00Z","updated_at":"2026-11-20T12:05:00Z"}`, id, status)
}

func TestCreateTemplateDefaultsTheBaseAndStartsABuild(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["POST /v2/sandboxes/templates"] = reply(201, `{"template":`+tplJSON("t-1", godo.SandboxTemplateStatusPending)+`}`)

	tpl, err := c.CreateTemplate(context.Background(), &sandbox.TemplateParams{Name: "my-template", SourceOCIRef: "r/i:v1"})
	if err != nil {
		t.Fatal(err)
	}
	if tpl.ID() != "t-1" || tpl.Name() != "my-template" || tpl.Status() != godo.SandboxTemplateStatusPending {
		t.Errorf("template = %+v", tpl.Info)
	}
	body := a.bodies[0]
	if body["name"] != "my-template" || body["base_template"] != "coding-base" || body["source_oci_ref"] != "r/i:v1" {
		t.Errorf("body = %+v", body)
	}

	for _, p := range []*sandbox.TemplateParams{nil, {SourceOCIRef: "r"}, {Name: "n"}, {Name: " ", SourceOCIRef: "r"}} {
		if _, err := c.CreateTemplate(context.Background(), p); err == nil {
			t.Errorf("%+v returned no error", p)
		}
	}
}

func TestGetUpdateAndDeleteTemplate(t *testing.T) {
	a, c := newAPI(t)
	ctx := context.Background()
	a.handlers["GET /v2/sandboxes/templates/t-1"] = reply(200, `{"template":`+tplJSON("t-1", godo.SandboxTemplateStatusReady)+`}`)
	a.handlers["PUT /v2/sandboxes/templates/t-1"] = reply(200, `{"template":`+tplJSON("t-1", godo.SandboxTemplateStatusBuilding)+`}`)
	a.handlers["DELETE /v2/sandboxes/templates/t-1"] = reply(200, `{}`)

	tpl, err := c.GetTemplate(ctx, "t-1")
	if err != nil || tpl.Status() != godo.SandboxTemplateStatusReady {
		t.Fatalf("GetTemplate = %+v, %v", tpl, err)
	}
	if err := tpl.Update(ctx, &sandbox.TemplateUpdate{SourceOCIRef: "r/i:v2"}); err != nil || tpl.Status() != godo.SandboxTemplateStatusBuilding {
		t.Fatalf("Update: status %s, err %v", tpl.Status(), err)
	}
	body := a.bodies[len(a.bodies)-1]
	if body["source_oci_ref"] != "r/i:v2" || body["base_template"] != nil {
		t.Errorf("update body = %+v", body)
	}
	if err := tpl.Update(ctx, &sandbox.TemplateUpdate{}); err == nil {
		t.Error("an empty update returned no error")
	}
	if deleted, err := tpl.Delete(ctx); err != nil || !deleted {
		t.Fatalf("Delete = %v, %v", deleted, err)
	}
}

func TestDeleteTemplateReportsFalseWhenGone(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["DELETE /v2/sandboxes/templates/gone"] = reply(404, `{"id":"not_found","message":"gone"}`)
	deleted, err := c.DeleteTemplate(context.Background(), "gone")
	if err != nil || deleted {
		t.Errorf("DeleteTemplate = %v, %v; expected false, nil", deleted, err)
	}
}

func TestListTemplatesPagesAndAll(t *testing.T) {
	a, c := newAPI(t)
	var queries []string
	a.handlers["GET /v2/sandboxes/templates"] = func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.RawQuery)
		if r.URL.Query().Get("page_token") == "" {
			fmt.Fprint(w, `{"templates":[`+tplJSON("a", godo.SandboxTemplateStatusReady)+`],"next_page_token":"t2"}`)
			return
		}
		fmt.Fprint(w, `{"templates":[`+tplJSON("b", godo.SandboxTemplateStatusReady)+`]}`)
	}
	pages := c.ListTemplates(&sandbox.ListTemplatesParams{Limit: 1})
	var ids []string
	for tpl, err := range pages.All(context.Background()) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, tpl.ID())
	}
	if strings.Join(ids, ",") != "a,b" || pages.HasNext() {
		t.Errorf("ids = %v, HasNext = %v", ids, pages.HasNext())
	}
	if !strings.Contains(queries[0], "page_size=1") || !strings.Contains(queries[1], "page_token=t2") {
		t.Errorf("queries = %v", queries)
	}
}

func TestTemplateBuilds(t *testing.T) {
	a, c := newAPI(t)
	ctx := context.Background()
	a.handlers["GET /v2/sandboxes/templates/t-1"] = reply(200, `{"template":`+tplJSON("t-1", godo.SandboxTemplateStatusReady)+`}`)
	a.handlers["GET /v2/sandboxes/templates/t-1/builds"] = reply(200,
		`{"builds":[{"build_id":"b-1","template_id":"t-1","status":"TEMPLATE_BUILD_STATUS_SUCCEEDED"}],"next_page_token":""}`)
	a.handlers["GET /v2/sandboxes/templates/t-1/builds/b-1"] = reply(200,
		`{"build":{"build_id":"b-1","template_id":"t-1","status":"TEMPLATE_BUILD_STATUS_SUCCEEDED"}}`)

	tpl, err := c.GetTemplate(ctx, "t-1")
	if err != nil {
		t.Fatal(err)
	}
	page, err := tpl.Builds(0).NextItems(ctx)
	if err != nil || len(page) != 1 || page[0].ID != "b-1" {
		t.Fatalf("Builds = %+v, %v", page, err)
	}
	build, err := tpl.Build(ctx, "b-1")
	if err != nil || build.Status != godo.SandboxTemplateBuildStatusSucceeded {
		t.Errorf("Build = %+v, %v", build, err)
	}
}

func TestWaitUntilReadyPollsABuildingTemplate(t *testing.T) {
	a, c := newAPI(t)
	n := 0
	a.handlers["GET /v2/sandboxes/templates/t-1"] = func(w http.ResponseWriter, r *http.Request) {
		n++
		status := godo.SandboxTemplateStatusBuilding
		if n >= 3 {
			status = godo.SandboxTemplateStatusReady
		}
		fmt.Fprint(w, `{"template":`+tplJSON("t-1", status)+`}`)
	}
	ctx := context.Background()
	tpl, err := c.GetTemplate(ctx, "t-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := tpl.WaitUntilReady(ctx, &sandbox.WaitOptions{PollInterval: time.Millisecond}); err != nil {
		t.Fatal(err)
	}
	if tpl.Status() != godo.SandboxTemplateStatusReady || n != 3 {
		t.Errorf("status %s after %d reads", tpl.Status(), n)
	}
}

func TestWaitUntilReadyExplainsAFailedBuild(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["GET /v2/sandboxes/templates/t-1"] = reply(200, `{"template":`+tplJSON("t-1", godo.SandboxTemplateStatusFailed)+`}`)
	a.handlers["GET /v2/sandboxes/templates/t-1/builds"] = reply(200, `{"builds":[`+
		`{"build_id":"old","template_id":"t-1","status":"TEMPLATE_BUILD_STATUS_FAILED","error":"old error","updated_at":"2026-11-20T10:00:00Z"},`+
		`{"build_id":"new","template_id":"t-1","status":"TEMPLATE_BUILD_STATUS_FAILED","error":"image not found","updated_at":"2026-11-20T12:00:00Z"}]}`)
	ctx := context.Background()
	tpl, _ := c.GetTemplate(ctx, "t-1")
	err := tpl.WaitUntilReady(ctx, nil)
	if err == nil || !strings.Contains(err.Error(), "failed to build: image not found") {
		t.Errorf("err = %v", err)
	}
}

func TestWaitUntilReadyTimesOut(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["GET /v2/sandboxes/templates/t-1"] = reply(200, `{"template":`+tplJSON("t-1", godo.SandboxTemplateStatusBuilding)+`}`)
	ctx := context.Background()
	tpl, _ := c.GetTemplate(ctx, "t-1")
	err := tpl.WaitUntilReady(ctx, &sandbox.WaitOptions{Timeout: 30 * time.Millisecond, PollInterval: time.Millisecond})
	if err == nil || !strings.Contains(err.Error(), "not ready in time") {
		t.Errorf("err = %v", err)
	}
}
