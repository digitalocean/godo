package godo

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// The tests in this file call every exported method of every service reachable
// from Client (and AgentInferenceClient) through reflection, so that a path
// building change is checked against the whole API surface rather than a few
// hand-picked methods. Requests are captured by recordingTransport and never
// leave the process.

const ordinaryPathValue = "a1:b2@c3.d-e_f~g"

type recordingTransport struct {
	mu   sync.Mutex
	reqs []*http.Request
}

func (rt *recordingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Body != nil {
		io.Copy(io.Discard, r.Body)
		r.Body.Close()
	}
	rt.mu.Lock()
	rt.reqs = append(rt.reqs, r)
	rt.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader("{}")),
		Request:    r,
	}, nil
}

func (rt *recordingTransport) take() []*http.Request {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	reqs := rt.reqs
	rt.reqs = nil
	return reqs
}

type serviceMethod struct {
	name string
	fn   reflect.Value
}

// collectServiceMethods returns every exported method of the services held by
// v's exported fields, descending into nested service structs such as
// ActionGateway.Tools and Chat.Completions.
func collectServiceMethods(prefix string, v reflect.Value, seen map[uintptr]bool) []serviceMethod {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() || seen[v.Pointer()] {
			return nil
		}
		seen[v.Pointer()] = true
	}
	var out []serviceMethod
	for i := 0; i < v.NumMethod(); i++ {
		out = append(out, serviceMethod{name: prefix + "." + v.Type().Method(i).Name, fn: v.Method(i)})
	}
	s := reflect.Indirect(v)
	if s.Kind() != reflect.Struct {
		return out
	}
	for i := 0; i < s.NumField(); i++ {
		f := s.Type().Field(i)
		if !f.IsExported() || !isGodoServiceType(f.Type) {
			continue
		}
		fv := s.Field(i)
		switch fv.Kind() {
		case reflect.Interface:
			if fv.IsNil() {
				continue
			}
			out = append(out, collectServiceMethods(prefix+"."+f.Name, fv.Elem(), seen)...)
			if fv.Elem().Kind() == reflect.Ptr {
				// Only the methods declared by the interface are API surface.
				out = filterInterfaceMethods(out, prefix+"."+f.Name, fv.Type())
			}
		case reflect.Ptr:
			if fv.Type().Elem().Kind() == reflect.Struct {
				out = append(out, collectServiceMethods(prefix+"."+f.Name, fv, seen)...)
			}
		}
	}
	return out
}

// isGodoServiceType excludes fields such as Client.HTTPClient and
// Client.BaseURL, whose methods are not API calls.
func isGodoServiceType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.PkgPath() == reflect.TypeOf(Client{}).PkgPath()
}

func filterInterfaceMethods(ms []serviceMethod, prefix string, iface reflect.Type) []serviceMethod {
	keep := ms[:0]
	for _, m := range ms {
		rest := strings.TrimPrefix(m.name, prefix+".")
		if rest == m.name || strings.Contains(rest, ".") {
			keep = append(keep, m)
			continue
		}
		if _, ok := iface.MethodByName(rest); ok {
			keep = append(keep, m)
		}
	}
	return keep
}

type requestHarness struct {
	transport *recordingTransport
	methods   []serviceMethod
}

func newRequestHarness(t *testing.T) *requestHarness {
	t.Helper()
	rt := &recordingTransport{}
	hc := &http.Client{Transport: rt}
	c := NewClient(hc)
	methods := collectServiceMethods("Client", reflect.ValueOf(c), map[uintptr]bool{})

	agent, err := NewAgentInferenceClient("https://agent.example.com", "key", SetAgentHTTPClient(hc))
	if err != nil {
		t.Fatalf("NewAgentInferenceClient: %v", err)
	}
	methods = append(methods, collectServiceMethods("AgentInferenceClient", reflect.ValueOf(agent), map[uintptr]bool{})...)

	// Client's own helpers are not service methods.
	keep := methods[:0]
	for _, m := range methods {
		if strings.Count(m.name, ".") > 1 || strings.HasPrefix(m.name, "AgentInferenceClient.") {
			keep = append(keep, m)
		}
	}
	sort.Slice(keep, func(i, j int) bool { return keep[i].name < keep[j].name })
	return &requestHarness{transport: rt, methods: keep}
}

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	readerValue = reflect.ValueOf(strings.NewReader("x"))
)

// argValue builds an argument of type t. String parameters get value; when
// fill is set, string fields of struct arguments get value too.
func argValue(ctx context.Context, t reflect.Type, value string, fill bool) reflect.Value {
	if t == contextType {
		return reflect.ValueOf(ctx)
	}
	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(value).Convert(t)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(1).Convert(t)
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(1.0).Convert(t)
	case reflect.Interface:
		if readerValue.Type().Implements(t) {
			return readerValue
		}
	case reflect.Ptr:
		p := reflect.New(t.Elem())
		if t.Elem().Kind() == reflect.Struct {
			fillStruct(p.Elem(), value, fill)
		} else if t.Elem().Kind() != reflect.Ptr {
			p.Elem().Set(argValue(ctx, t.Elem(), value, fill))
		}
		return p
	case reflect.Struct:
		s := reflect.New(t).Elem()
		fillStruct(s, value, fill)
		return s
	}
	return reflect.Zero(t)
}

func fillStruct(s reflect.Value, value string, fill bool) {
	ordinary := ordinaryPathValue
	if fill {
		ordinary = value
	}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if !s.Type().Field(i).IsExported() || !f.CanSet() {
			continue
		}
		switch f.Kind() {
		case reflect.String:
			f.Set(reflect.ValueOf(ordinary).Convert(f.Type()))
		case reflect.Int, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint32, reflect.Uint64:
			f.Set(reflect.ValueOf(1).Convert(f.Type()))
		case reflect.Interface:
			if readerValue.Type().Implements(f.Type()) {
				f.Set(readerValue)
			}
		}
	}
}

type callResult struct {
	reqs     []*http.Request
	err      error
	panicked interface{}
	timedOut bool
}

// call invokes m with string parameter at position target set to value (every
// string parameter when target is -1, none when value is ordinary).
func (h *requestHarness) call(m serviceMethod, target int, value string, fill bool) callResult {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ft := m.fn.Type()
	n := ft.NumIn()
	if ft.IsVariadic() {
		n--
	}
	args := make([]reflect.Value, n)
	for i := 0; i < n; i++ {
		v := ordinaryPathValue
		if target == -1 || i == target {
			v = value
		}
		args[i] = argValue(ctx, ft.In(i), v, fill)
	}

	done := make(chan callResult, 1)
	go func() {
		var res callResult
		defer func() {
			res.panicked = recover()
			done <- res
		}()
		out := m.fn.Call(args)
		if len(out) > 0 {
			if last := out[len(out)-1]; last.Type() == errorType && !last.IsNil() {
				res.err = last.Interface().(error)
			}
		}
	}()
	var res callResult
	select {
	case res = <-done:
	case <-time.After(3 * time.Second):
		res.timedOut = true
	}
	res.reqs = h.transport.take()
	return res
}

func stringParams(ft reflect.Type) []int {
	var idx []int
	for i := 0; i < ft.NumIn(); i++ {
		if ft.In(i).Kind() == reflect.String {
			idx = append(idx, i)
		}
	}
	return idx
}

func isPathArgError(err error) bool {
	var ae *ArgError
	return errors.As(err, &ae) && ae.arg == "path"
}

func TestAllServices_harnessCoverage(t *testing.T) {
	h := newRequestHarness(t)
	sent := 0
	for _, m := range h.methods {
		if len(h.call(m, -1, ordinaryPathValue, false).reqs) > 0 {
			sent++
		}
	}
	t.Logf("%d service methods, %d sent a request with ordinary arguments", len(h.methods), sent)
	// Guards against the harness silently exercising nothing, e.g. if services
	// move off Client.
	if len(h.methods) < 850 || sent < 800 {
		t.Fatalf("harness reached %d methods and %d requests; expected the full API surface", len(h.methods), sent)
	}
}

// Ordinary IDs, including characters real IDs use (":" in IPv6 addresses and
// fingerprints, "@" in user names, "." in topic names), must never be rejected
// by the path check or altered by escaping.
func TestAllServices_ordinaryArgumentsAccepted(t *testing.T) {
	h := newRequestHarness(t)
	for _, m := range h.methods {
		res := h.call(m, -1, ordinaryPathValue, true)
		if isPathArgError(res.err) {
			t.Errorf("%s rejected ordinary arguments: %v", m.name, res.err)
		}
		for _, r := range res.reqs {
			if err := checkRequestPath(r.URL); err != nil {
				t.Errorf("%s sent %s %s: %v", m.name, r.Method, r.URL, err)
			}
			if strings.Contains(r.URL.EscapedPath(), "%3A") || strings.Contains(r.URL.EscapedPath(), "%40") {
				t.Errorf("%s over-escaped %s", m.name, r.URL.EscapedPath())
			}
		}
	}
}

// Methods whose string parameter is a full URI by design.
var uriParamMethods = map[string]bool{
	"Client.DropletActions.GetByURI":        true,
	"Client.ImageActions.GetByURI":          true,
	"Client.BatchInference.UploadInputFile": true,
}

func TestAllServices_traversalCannotChangeRoute(t *testing.T) {
	h := newRequestHarness(t)
	attacks := []string{
		"../../v2/account/keys/999",
		"x/../../../v2/account/keys/999",
		"..",
		"%2e%2e%2f%2e%2e%2fv2%2faccount%2fkeys%2f999",
	}
	for _, m := range h.methods {
		params := stringParams(m.fn.Type())
		targets := append([]int{}, params...)
		targets = append(targets, -1)
		for _, attack := range attacks {
			for _, target := range targets {
				res := h.call(m, target, attack, target == -1)
				for _, r := range res.reqs {
					p := r.URL.EscapedPath()
					if strings.Contains(p, "/account/keys/999") {
						t.Errorf("%s(param %d = %q) reached %s %s", m.name, target, attack, r.Method, r.URL)
					}
					if err := checkRequestPath(r.URL); err != nil {
						t.Errorf("%s(param %d = %q) sent %s %s: %v", m.name, target, attack, r.Method, r.URL, err)
					}
				}
			}
		}
	}
}

func TestAllServices_segmentInjectionEscaped(t *testing.T) {
	h := newRequestHarness(t)
	const attack = "x/y?inj=1#frag"
	for _, m := range h.methods {
		if uriParamMethods[m.name] {
			continue
		}
		targets := append(stringParams(m.fn.Type()), -1)
		for _, target := range targets {
			res := h.call(m, target, attack, target == -1)
			for _, r := range res.reqs {
				if strings.Contains(r.URL.EscapedPath(), "x/y") {
					t.Errorf("%s(param %d) put an unescaped \"/\" in the path: %s", m.name, target, r.URL.EscapedPath())
				}
				if strings.Contains(r.URL.RawQuery, "inj=1") {
					t.Errorf("%s(param %d) injected a query parameter: %s", m.name, target, r.URL.RawQuery)
				}
				if r.URL.Fragment != "" {
					t.Errorf("%s(param %d) produced a fragment: %s", m.name, target, r.URL)
				}
			}
		}
	}
}

// Empty IDs must not turn a request on a resource into a request on its
// collection, e.g. DELETE /v2/databases/ instead of DELETE /v2/databases/{id}.
func TestAllServices_emptyIDDoesNotReachCollection(t *testing.T) {
	h := newRequestHarness(t)
	for _, m := range h.methods {
		if uriParamMethods[m.name] {
			continue
		}
		for _, target := range stringParams(m.fn.Type()) {
			res := h.call(m, target, "", false)
			for _, r := range res.reqs {
				if strings.HasSuffix(r.URL.Path, "/") || strings.Contains(r.URL.Path, "//") {
					t.Errorf("%s(param %d = \"\") sent %s %s", m.name, target, r.Method, r.URL)
				}
			}
		}
	}
}
