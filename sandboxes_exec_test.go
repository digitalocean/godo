package godo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
)

func TestSandboxes_StartExec(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/exec", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q", got)
		}
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		want := map[string]interface{}{
			"argv":            []interface{}{"bash", "-lc", "echo hello"},
			"workdir":         "/workspace",
			"env_vars":        map[string]interface{}{"A": "1"},
			"timeout_seconds": "30",
			"keep_stdin_open": true,
			"stdin":           b64("hi\n"),
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("request body = %+v, expected %+v", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		// One envelope split without a newline, a keepalive, and an exit
		// without exit_code (protobuf JSON omits zero).
		fmt.Fprint(w, `{"result":{"started":{"exec_id":"ex-1"}}}`+
			`{"result":{"keepalive":{}}}`+"\n"+
			fmt.Sprintf(`{"result":{"stdout":%q}}`, b64("hello\n"))+"\n"+
			fmt.Sprintf(`{"result":{"stderr":%q}}`, b64("oops\n"))+"\n"+
			`{"result":{"exit":{"duration_ms":"39"}}}`+"\n")
	})

	stream, _, err := client.Sandboxes.StartExec(ctx, "sb-1", &SandboxExecStartRequest{
		Argv: []string{"bash", "-lc", "echo hello"}, Workdir: "/workspace", EnvVars: map[string]string{"A": "1"},
		TimeoutSeconds: 30, KeepStdinOpen: true, Stdin: []byte("hi\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	var kinds []string
	for {
		ev, err := stream.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case ev.Started != nil:
			kinds = append(kinds, "started:"+ev.Started.ExecID)
		case ev.Keepalive != nil:
			kinds = append(kinds, "keepalive")
		case ev.Stdout != nil:
			kinds = append(kinds, "stdout:"+string(ev.Stdout))
		case ev.Stderr != nil:
			kinds = append(kinds, "stderr:"+string(ev.Stderr))
		case ev.Exit != nil:
			kinds = append(kinds, fmt.Sprintf("exit:%d:%d", ev.Exit.ExitCode, ev.Exit.DurationMS))
		}
	}
	want := []string{"started:ex-1", "keepalive", "stdout:hello\n", "stderr:oops\n", "exit:0:39"}
	if !reflect.DeepEqual(kinds, want) {
		t.Errorf("events = %v, expected %v", kinds, want)
	}
}

func TestSandboxes_StartExec_ptyAndValidation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/exec", func(w http.ResponseWriter, r *http.Request) {
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		want := map[string]interface{}{"pty": map[string]interface{}{"rows": float64(40), "cols": float64(120), "term": "xterm"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("request body = %+v, expected %+v", got, want)
		}
		fmt.Fprint(w, `{"result":{"started":{"exec_id":"ex-1"}}}`+"\n")
	})
	stream, _, err := client.Sandboxes.StartExec(ctx, "sb-1", &SandboxExecStartRequest{PTY: &SandboxExecPTY{Rows: 40, Cols: 120, Term: "xterm"}})
	if err != nil {
		t.Fatal(err)
	}
	stream.Close()

	if _, _, err := client.Sandboxes.StartExec(ctx, "sb-1", &SandboxExecStartRequest{}); err == nil {
		t.Error("an empty argv without pty returned no error")
	}
	if _, _, err := client.Sandboxes.StartExec(ctx, "", &SandboxExecStartRequest{Argv: []string{"ls"}}); err == nil {
		t.Error("an empty sandbox id returned no error")
	}
}

func TestSandboxes_StartExec_errors(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/exec", func(w http.ResponseWriter, r *http.Request) {
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		if got["pty"] != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"id":"bad_request","message":"this sandbox's template image does not support interactive stdin or PTY exec; create a new sandbox","request_id":"req-1"}`)
			return
		}
		fmt.Fprint(w, fmt.Sprintf(`{"result":{"stdout":%q}}`, b64("a"))+"\n"+`{"error":{"http_code":500,"message":"gateway failed"}}`+"\n")
	})

	_, _, err := client.Sandboxes.StartExec(ctx, "sb-1", &SandboxExecStartRequest{PTY: &SandboxExecPTY{}})
	if apiErr, ok := err.(*ErrorResponse); !ok || apiErr.Response.StatusCode != http.StatusBadRequest || apiErr.RequestID != "req-1" {
		t.Errorf("StartExec returned %#v, expected a 400 ErrorResponse", err)
	}

	stream, _, err := client.Sandboxes.StartExec(ctx, "sb-1", &SandboxExecStartRequest{Argv: []string{"ls"}})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if _, err := stream.Next(); err != nil {
		t.Fatal(err)
	}
	_, err = stream.Next()
	if se, ok := err.(*SandboxCommandStreamError); !ok || se.HTTPCode != 500 || se.Error() != "gateway failed" {
		t.Errorf("Next returned %#v, expected the stream error", err)
	}
}

func TestSandboxes_WriteExecStdinAndResize(t *testing.T) {
	setup()
	defer teardown()

	var bodies []map[string]interface{}
	mux.HandleFunc("/v2/sandboxes/sb-1/exec/ex-1/stdin", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		bodies = append(bodies, got)
		fmt.Fprint(w, `{"offset":"3"}`)
	})
	mux.HandleFunc("/v2/sandboxes/sb-1/exec/ex-1/resize", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		bodies = append(bodies, got)
		fmt.Fprint(w, `{}`)
	})

	off, _, err := client.Sandboxes.WriteExecStdin(ctx, "sb-1", "ex-1", &SandboxExecStdinRequest{Data: []byte("hi\n")})
	if err != nil || off != 3 {
		t.Fatalf("WriteExecStdin returned %d, %v", off, err)
	}
	if _, _, err := client.Sandboxes.WriteExecStdin(ctx, "sb-1", "ex-1", &SandboxExecStdinRequest{Offset: 3, EOF: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Sandboxes.ResizeExec(ctx, "sb-1", "ex-1", 40, 120); err != nil {
		t.Fatal(err)
	}
	want := []map[string]interface{}{
		{"data": b64("hi\n"), "offset": "0"},
		{"offset": "3", "eof": true},
		{"rows": float64(40), "cols": float64(120)},
	}
	if !reflect.DeepEqual(bodies, want) {
		t.Errorf("bodies = %+v, expected %+v", bodies, want)
	}
	if _, _, err := client.Sandboxes.WriteExecStdin(ctx, "sb-1", "", &SandboxExecStdinRequest{}); err == nil {
		t.Error("an empty exec id returned no error")
	}
	if _, err := client.Sandboxes.ResizeExec(ctx, "sb-1", "", 1, 1); err == nil {
		t.Error("an empty exec id returned no error")
	}
}
