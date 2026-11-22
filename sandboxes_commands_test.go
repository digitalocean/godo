package godo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
)

func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

func TestSandboxes_StartCommand(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/commands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		want := map[string]interface{}{
			"cmd":             "sleep",
			"args":            []interface{}{"60"},
			"cwd":             "/workspace",
			"env_vars":        map[string]interface{}{"A": "1"},
			"stdin":           true,
			"timeout_seconds": "30",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("request body = %+v, expected %+v", got, want)
		}
		fmt.Fprint(w, `{"pid":42}`)
	})

	pid, _, err := client.Sandboxes.StartCommand(ctx, "sb-1", &SandboxCommandStartRequest{
		Cmd: "sleep", Args: []string{"60"}, Cwd: "/workspace", EnvVars: map[string]string{"A": "1"},
		Stdin: true, TimeoutSeconds: 30,
	})
	if err != nil || pid != 42 {
		t.Fatalf("StartCommand returned %d, %v", pid, err)
	}
}

func TestSandboxes_StartCommand_omitsUnsetFieldsAndValidates(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/commands", func(w http.ResponseWriter, r *http.Request) {
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		if !reflect.DeepEqual(got, map[string]interface{}{"cmd": "ls"}) {
			t.Errorf("request body = %+v, expected only cmd", got)
		}
		fmt.Fprint(w, `{"pid":1}`)
	})
	if _, _, err := client.Sandboxes.StartCommand(ctx, "sb-1", &SandboxCommandStartRequest{Cmd: "ls"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.Sandboxes.StartCommand(ctx, "sb-1", &SandboxCommandStartRequest{}); err == nil {
		t.Error("an empty command returned no error")
	}
	if _, _, err := client.Sandboxes.StartCommand(ctx, "", &SandboxCommandStartRequest{Cmd: "ls"}); err == nil {
		t.Error("an empty sandbox id returned no error")
	}
}

func TestSandboxes_ListCommands(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/commands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"commands":[{"pid":42,"cmd":"bash","cwd":"/workspace"},{"pid":43,"cmd":"sleep"}]}`)
	})
	got, _, err := client.Sandboxes.ListCommands(ctx, "sb-1")
	want := []SandboxCommand{{PID: 42, Cmd: "bash", Cwd: "/workspace"}, {PID: 43, Cmd: "sleep"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("ListCommands returned %+v, %v; expected %+v", got, err, want)
	}
}

func TestSandboxes_StreamCommand(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/commands/42/stream", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"result":{"started":{"pid":42}}}`+"\n"+
			`{"result":{"keepalive":{}}}`+"\n"+
			fmt.Sprintf(`{"result":{"stdout":%q}}`, b64("hi\n"))+"\n"+
			fmt.Sprintf(`{"result":{"stderr":%q}}`, b64("warn"))+"\n"+
			`{"result":{"exit":{"exit_code":3,"status":"exited"}}}`+"\n")
	})

	stream, _, err := client.Sandboxes.StreamCommand(ctx, "sb-1", 42)
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
			kinds = append(kinds, fmt.Sprintf("started:%d", ev.Started.PID))
		case ev.Keepalive != nil:
			kinds = append(kinds, "keepalive")
		case ev.Stdout != nil:
			kinds = append(kinds, "stdout:"+string(ev.Stdout))
		case ev.Stderr != nil:
			kinds = append(kinds, "stderr:"+string(ev.Stderr))
		case ev.Exit != nil:
			kinds = append(kinds, fmt.Sprintf("exit:%d:%s", ev.Exit.ExitCode, ev.Exit.Status))
		}
	}
	want := []string{"started:42", "keepalive", "stdout:hi\n", "stderr:warn", "exit:3:exited"}
	if !reflect.DeepEqual(kinds, want) {
		t.Errorf("events = %v, expected %v", kinds, want)
	}
}

func TestSandboxes_StreamCommand_errors(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/commands/99/stream", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"id":"not_found","message":"not found","request_id":"req-1"}`)
	})
	mux.HandleFunc("/v2/sandboxes/sb-1/commands/42/stream", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"result":{"started":{"pid":42}}}`+"\n"+`{"error":{"http_code":500,"message":"gateway failed"}}`+"\n")
	})

	if _, _, err := client.Sandboxes.StreamCommand(ctx, "sb-1", 99); err == nil {
		t.Fatal("an unknown pid returned no error")
	} else if apiErr, ok := err.(*ErrorResponse); !ok || apiErr.Response.StatusCode != http.StatusNotFound || apiErr.RequestID != "req-1" {
		t.Errorf("StreamCommand returned %#v, expected a 404 ErrorResponse", err)
	}

	stream, _, err := client.Sandboxes.StreamCommand(ctx, "sb-1", 42)
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

func TestSandboxes_SignalAndStdin(t *testing.T) {
	setup()
	defer teardown()

	var bodies []map[string]interface{}
	record := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		bodies = append(bodies, got)
		fmt.Fprint(w, `{}`)
	}
	mux.HandleFunc("/v2/sandboxes/sb-1/commands/42/signal", record)
	mux.HandleFunc("/v2/sandboxes/sb-1/commands/42/stdin", record)

	if _, err := client.Sandboxes.SignalCommand(ctx, "sb-1", 42, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Sandboxes.SignalCommand(ctx, "sb-1", 42, SandboxCommandSignalKill); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Sandboxes.WriteCommandStdin(ctx, "sb-1", 42, &SandboxCommandStdinRequest{Data: []byte("hello\n"), EOF: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Sandboxes.WriteCommandStdin(ctx, "sb-1", 42, &SandboxCommandStdinRequest{EOF: true}); err != nil {
		t.Fatal(err)
	}
	want := []map[string]interface{}{
		{"signal": "SANDBOX_COMMAND_SIGNAL_SIGTERM"},
		{"signal": "SANDBOX_COMMAND_SIGNAL_SIGKILL"},
		{"data": b64("hello\n"), "eof": true},
		{"eof": true},
	}
	if !reflect.DeepEqual(bodies, want) {
		t.Errorf("bodies = %+v, expected %+v", bodies, want)
	}
}
