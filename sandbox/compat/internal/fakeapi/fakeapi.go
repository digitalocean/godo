// Package fakeapi is a small fake of the DigitalOcean Sandboxes API for tests.
package fakeapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// API serves handlers by "METHOD /path" and records each request.
type API struct {
	URL      string
	Handlers map[string]http.HandlerFunc

	mu     sync.Mutex
	calls  []string
	bodies []map[string]any
	auth   string
}

// New starts a fake API that is stopped when the test ends.
func New(t testing.TB) *API {
	t.Helper()
	a := &API{Handlers: map[string]http.HandlerFunc{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		a.mu.Lock()
		a.calls = append(a.calls, key)
		a.bodies = append(a.bodies, body)
		a.auth = r.Header.Get("Authorization")
		h, ok := a.Handlers[key]
		a.mu.Unlock()
		if !ok {
			t.Errorf("unexpected request %s", key)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	a.URL = srv.URL
	return a
}

// Reply returns a handler that answers with a status and a JSON body.
func Reply(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}
}

// Calls returns the "METHOD /path" of every request so far.
func (a *API) Calls() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string{}, a.calls...)
}

// Body returns the JSON body of the last request for key, or nil.
func (a *API) Body(key string) map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := len(a.calls) - 1; i >= 0; i-- {
		if a.calls[i] == key {
			return a.bodies[i]
		}
	}
	return nil
}

// Auth returns the Authorization header of the last request.
func (a *API) Auth() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.auth
}

// Sandbox is the JSON of a sandbox.
func Sandbox(id, state, template string) string {
	return fmt.Sprintf(`{"sandbox_id":%q,"template":%q,"state":%q,"size_slug":"mars-2vcpu-4gb","created_at":"2026-11-20T12:00:00Z"}`, id, template, state)
}

// Template is the JSON of a template.
func Template(id, name, status string) string {
	return fmt.Sprintf(`{"template_id":%q,"name":%q,"status":%q,"spec":{"base_template":"coding-base","image":{"registry":"docker.io","repository":"library/python","tag":"3.12"}},"created_at":"2026-11-20T12:00:00Z","updated_at":"2026-11-20T12:05:00Z"}`, id, name, status)
}

// Snapshot is the JSON of a snapshot.
func Snapshot(id, state string) string {
	return fmt.Sprintf(`{"snapshot_id":%q,"sandbox_id":"sb-1","name":"snap","state":%q,"created_at":"2026-11-20T12:00:00Z"}`, id, state)
}

// Stream is the body of a command's attach stream: an exit event after the
// given output chunks, each written as a stdout event.
func Stream(exitCode int, stdout ...string) string {
	var out string
	out += `{"result":{"started":{"pid":42}}}` + "\n"
	for _, s := range stdout {
		out += fmt.Sprintf(`{"result":{"stdout":%q}}`, b64(s)) + "\n"
	}
	out += fmt.Sprintf(`{"result":{"exit":{"exit_code":%d,"status":"exited"}}}`, exitCode) + "\n"
	return out
}
