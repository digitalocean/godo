package godo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

const sandboxFileEntryJSON = `{"name":"hello.txt","type":"FILE_ENTRY_TYPE_FILE","path":"/workspace/hello.txt","size":"12","mode":420,"permissions":"-rw-r--r--","owner":"user","group":"user","modified_time":"2026-03-25T12:00:00Z"}`

func sandboxFileEntryWant() SandboxFileEntry {
	return SandboxFileEntry{
		Name: "hello.txt", Type: SandboxFileEntryTypeFile, Path: "/workspace/hello.txt",
		Size: 12, Mode: 420, Permissions: "-rw-r--r--", Owner: "user", Group: "user",
		ModifiedTime: Timestamp{time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)},
	}
}

func sandboxDownloadBody(payload []byte) []byte {
	sum := sha256.Sum256(payload)
	return append(append([]byte{}, payload...), []byte("DOWSSHA1"+hex.EncodeToString(sum[:])+"\n")...)
}

func shortSandboxFileRetries(t *testing.T) {
	old := sandboxFileRetryDelays
	sandboxFileRetryDelays = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}
	t.Cleanup(func() { sandboxFileRetryDelays = old })
}

func TestSandboxes_UploadFile(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/files/upload", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testFormValues(t, r, values{"path": "/workspace/hello.txt"})
		if ct := r.Header.Get("Content-Type"); ct != "application/octet-stream" {
			t.Errorf("Content-Type = %q", ct)
		}
		if sum := r.Header.Get("X-Content-Sha256"); sum != "abc123" {
			t.Errorf("X-Content-Sha256 = %q", sum)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "hello world\n" {
			t.Errorf("body = %q", body)
		}
		fmt.Fprint(w, `{"path":"/workspace/hello.txt","bytes_written":"12"}`)
	})

	got, _, err := client.Sandboxes.UploadFile(ctx, "sb-1", &SandboxFileUploadRequest{
		Path: "/workspace/hello.txt", Body: strings.NewReader("hello world\n"), SHA256: "abc123",
	})
	if err != nil || !reflect.DeepEqual(got, &SandboxFileUpload{Path: "/workspace/hello.txt", BytesWritten: 12}) {
		t.Fatalf("UploadFile returned %+v, %v", got, err)
	}
}

func TestSandboxes_UploadArchiveWithoutChecksum(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/files/upload", func(w http.ResponseWriter, r *http.Request) {
		testFormValues(t, r, values{"path": "/workspace/out", "is_archive": "true"})
		if _, ok := r.Header["X-Content-Sha256"]; ok {
			t.Error("X-Content-Sha256 was sent without a checksum")
		}
		fmt.Fprint(w, `{"path":"/workspace/out","bytes_written":"3"}`)
	})

	got, _, err := client.Sandboxes.UploadFile(ctx, "sb-1", &SandboxFileUploadRequest{
		Path: "/workspace/out", Body: strings.NewReader("tar"), IsArchive: true,
	})
	if err != nil || got.BytesWritten != 3 {
		t.Fatalf("UploadFile returned %+v, %v", got, err)
	}
}

func TestSandboxes_UploadFileValidates(t *testing.T) {
	setup()
	defer teardown()

	for _, tc := range []struct {
		id  string
		req *SandboxFileUploadRequest
	}{
		{"", &SandboxFileUploadRequest{Path: "/a"}},
		{"sb-1", nil},
		{"sb-1", &SandboxFileUploadRequest{}},
		{"sb-1", &SandboxFileUploadRequest{Path: "/a", Body: strings.NewReader(strings.Repeat("x", SandboxFileMaxBytes+1))}},
	} {
		if _, _, err := client.Sandboxes.UploadFile(ctx, tc.id, tc.req); err == nil {
			t.Errorf("UploadFile(%q, %+v) returned no error", tc.id, tc.req)
		}
	}
}

func TestSandboxes_UploadFileRetries503(t *testing.T) {
	setup()
	defer teardown()
	shortSandboxFileRetries(t)

	calls := 0
	mux.HandleFunc("/v2/sandboxes/sb-1/files/upload", func(w http.ResponseWriter, r *http.Request) {
		calls++
		body, _ := io.ReadAll(r.Body)
		if string(body) != "data" {
			t.Errorf("attempt %d body = %q", calls, body)
		}
		if calls < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"id":"service_unavailable","message":"too many transfers"}`)
			return
		}
		fmt.Fprint(w, `{"path":"/workspace/a","bytes_written":"4"}`)
	})

	got, _, err := client.Sandboxes.UploadFile(ctx, "sb-1", &SandboxFileUploadRequest{Path: "/workspace/a", Body: strings.NewReader("data")})
	if err != nil || got.BytesWritten != 4 || calls != 3 {
		t.Fatalf("UploadFile returned %+v, %v after %d calls", got, err, calls)
	}
}

func TestSandboxes_DownloadFileGivesUpAfterThreeRetries(t *testing.T) {
	setup()
	defer teardown()
	shortSandboxFileRetries(t)

	calls := 0
	mux.HandleFunc("/v2/sandboxes/sb-1/files/download", func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	_, resp, err := client.Sandboxes.DownloadFile(ctx, "sb-1", &SandboxFileDownloadRequest{Path: "/workspace/a"})
	if err == nil || resp == nil || resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("DownloadFile returned %v, %v", resp, err)
	}
	if calls != 4 {
		t.Errorf("calls = %d, expected 4", calls)
	}
}

func TestSandboxes_TransferDoesNotRetryOtherErrors(t *testing.T) {
	setup()
	defer teardown()
	shortSandboxFileRetries(t)

	calls := 0
	mux.HandleFunc("/v2/sandboxes/sb-1/files/upload", func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	})
	_, resp, err := client.Sandboxes.UploadFile(ctx, "sb-1", &SandboxFileUploadRequest{Path: "/a", Body: strings.NewReader("x")})
	if err == nil || resp.StatusCode != http.StatusRequestEntityTooLarge || calls != 1 {
		t.Fatalf("UploadFile returned %v after %d calls", err, calls)
	}
}

func TestSandboxes_TransferLeavesRetriesToRetryConfig(t *testing.T) {
	setup()
	defer teardown()
	shortSandboxFileRetries(t)
	client.RetryConfig.RetryMax = 1 // the HTTP client itself is left plain here

	calls := 0
	mux.HandleFunc("/v2/sandboxes/sb-1/files/download", func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	if _, _, err := client.Sandboxes.DownloadFile(ctx, "sb-1", &SandboxFileDownloadRequest{Path: "/a"}); err == nil || calls != 1 {
		t.Fatalf("DownloadFile returned %v after %d calls", err, calls)
	}
}

func TestSandboxes_DownloadFile(t *testing.T) {
	setup()
	defer teardown()

	payload := []byte("hello world\n")
	mux.HandleFunc("/v2/sandboxes/sb-1/files/download", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{"path": "/workspace/dir", "as_archive": "true"})
		if a := r.Header.Get("Accept"); a != "application/octet-stream" {
			t.Errorf("Accept = %q", a)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Workspace-Is-Archive", "true")
		w.Header().Set("X-Workspace-Size-Bytes", "12")
		_, _ = w.Write(sandboxDownloadBody(payload))
	})

	got, resp, err := client.Sandboxes.DownloadFile(ctx, "sb-1", &SandboxFileDownloadRequest{Path: "/workspace/dir", AsArchive: true})
	if err != nil || string(got) != string(payload) {
		t.Fatalf("DownloadFile returned %q, %v", got, err)
	}
	if resp.Header.Get("X-Workspace-Is-Archive") != "true" || resp.Header.Get("X-Workspace-Size-Bytes") != "12" {
		t.Errorf("headers = %v", resp.Header)
	}
	for _, req := range []*SandboxFileDownloadRequest{nil, {}} {
		if _, _, err := client.Sandboxes.DownloadFile(ctx, "sb-1", req); err == nil {
			t.Errorf("DownloadFile(%+v) returned no error", req)
		}
	}
	if _, _, err := client.Sandboxes.DownloadFile(ctx, "", &SandboxFileDownloadRequest{Path: "/a"}); err == nil {
		t.Error("an empty sandbox id returned no error")
	}
}

func TestSandboxes_DownloadFileRejectsABadFooter(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/files/download", func(w http.ResponseWriter, r *http.Request) {
		body := sandboxDownloadBody([]byte("hello"))
		body[0] = 'j'
		_, _ = w.Write(body)
	})
	if _, _, err := client.Sandboxes.DownloadFile(ctx, "sb-1", &SandboxFileDownloadRequest{Path: "/a"}); err == nil ||
		!strings.Contains(err.Error(), "integrity check") {
		t.Fatalf("DownloadFile returned %v", err)
	}
}

func TestVerifySandboxDownload(t *testing.T) {
	good := sandboxDownloadBody([]byte("payload"))
	sum := sha256.Sum256([]byte("payload"))
	digest := hex.EncodeToString(sum[:])

	if got, err := verifySandboxDownload(good); err != nil || string(got) != "payload" {
		t.Errorf("good body = %q, %v", got, err)
	}
	if got, err := verifySandboxDownload(sandboxDownloadBody(nil)); err != nil || len(got) != 0 {
		t.Errorf("empty payload = %q, %v", got, err)
	}
	upper := []byte("payload" + "DOWSSHA1" + strings.ToUpper(digest) + "\n")
	if got, err := verifySandboxDownload(upper); err != nil || string(got) != "payload" {
		t.Errorf("upper-case digest = %q, %v", got, err)
	}

	for name, body := range map[string][]byte{
		"short body":   good[len(good)-72:],
		"no footer":    []byte(strings.Repeat("x", 100)),
		"bad magic":    []byte("payload" + "DOWSSHA2" + digest + "\n"),
		"bad newline":  []byte("payload" + "DOWSSHA1" + digest + "\r"),
		"non-hex":      []byte("payload" + "DOWSSHA1" + strings.Repeat("z", 64) + "\n"),
		"short digest": []byte("payload" + "DOWSSHA1" + digest[:63] + "\n"),
		"sha mismatch": []byte("Payload" + "DOWSSHA1" + digest + "\n"),
	} {
		if _, err := verifySandboxDownload(body); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestParseSandboxRetryAfter(t *testing.T) {
	for in, want := range map[string]time.Duration{"0": 0, "2": 2 * time.Second, "3600": sandboxFileMaxRetryAfter} {
		if d, ok := parseSandboxRetryAfter(in); !ok || d != want {
			t.Errorf("%q = %v, %v", in, d, ok)
		}
	}
	for _, in := range []string{"", "-1", "soon"} {
		if _, ok := parseSandboxRetryAfter(in); ok {
			t.Errorf("%q parsed", in)
		}
	}
	if d, ok := parseSandboxRetryAfter(time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)); !ok || d != 0 {
		t.Errorf("a past date = %v, %v", d, ok)
	}
}

func TestSandboxes_ListFiles(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/files/list", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		want := values{"path": "/workspace"}
		if r.URL.Query().Get("depth") != "" {
			want["depth"] = "2"
		}
		testFormValues(t, r, want)
		fmt.Fprintf(w, `{"entries":[%s]}`, sandboxFileEntryJSON)
	})

	got, _, err := client.Sandboxes.ListFiles(ctx, "sb-1", "/workspace", 2)
	if err != nil || !reflect.DeepEqual(got, []SandboxFileEntry{sandboxFileEntryWant()}) {
		t.Fatalf("ListFiles returned %+v, %v", got, err)
	}
	if _, _, err := client.Sandboxes.ListFiles(ctx, "sb-1", "/workspace", 0); err != nil {
		t.Fatalf("depth 0 returned %v", err)
	}
	if _, _, err := client.Sandboxes.ListFiles(ctx, "sb-1", "", 0); err == nil {
		t.Error("an empty path returned no error")
	}
	if _, _, err := client.Sandboxes.ListFiles(ctx, "sb-1", "/workspace", -1); err == nil {
		t.Error("a negative depth returned no error")
	}
}

func TestSandboxes_GetFileInfo(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/files/info", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if r.URL.Query().Get("path") == "/missing" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"id":"not_found","message":"no such file"}`)
			return
		}
		fmt.Fprintf(w, `{"entry":%s}`, sandboxFileEntryJSON)
	})

	got, _, err := client.Sandboxes.GetFileInfo(ctx, "sb-1", "/workspace/hello.txt")
	want := sandboxFileEntryWant()
	if err != nil || !reflect.DeepEqual(got, &want) {
		t.Fatalf("GetFileInfo returned %+v, %v", got, err)
	}
	if _, resp, err := client.Sandboxes.GetFileInfo(ctx, "sb-1", "/missing"); err == nil || resp.StatusCode != http.StatusNotFound {
		t.Fatalf("a missing path returned %v", err)
	}
}

func TestSandboxes_GetFileInfoSymlink(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/files/info", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"entry":{"name":"link","type":"FILE_ENTRY_TYPE_SYMLINK","path":"/workspace/link","symlink_target":"/workspace/hello.txt","modified_time":"2026-03-25T12:00:00Z"}}`)
	})
	got, _, err := client.Sandboxes.GetFileInfo(ctx, "sb-1", "/workspace/link")
	if err != nil || got.Type != SandboxFileEntryTypeSymlink || got.SymlinkTarget != "/workspace/hello.txt" || got.Size != 0 {
		t.Fatalf("GetFileInfo returned %+v, %v", got, err)
	}
}

func TestSandboxes_FileExists(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/files/exists", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if r.URL.Query().Get("path") == "/workspace/hello.txt" {
			fmt.Fprint(w, `{"exists":true}`)
			return
		}
		fmt.Fprint(w, `{}`)
	})

	if ok, _, err := client.Sandboxes.FileExists(ctx, "sb-1", "/workspace/hello.txt"); err != nil || !ok {
		t.Fatalf("FileExists = %v, %v", ok, err)
	}
	if ok, _, err := client.Sandboxes.FileExists(ctx, "sb-1", "/nope"); err != nil || ok {
		t.Fatalf("FileExists({}) = %v, %v", ok, err)
	}
	if _, _, err := client.Sandboxes.FileExists(ctx, "sb-1", ""); err == nil {
		t.Error("an empty path returned no error")
	}
}

func TestSandboxes_MakeDirectory(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/files/make-directory", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got map[string]string
		_ = json.NewDecoder(r.Body).Decode(&got)
		entry := `{"name":"outdir","type":"FILE_ENTRY_TYPE_DIRECTORY","path":"/workspace/outdir","size":"0","mode":493,"modified_time":"2026-03-25T12:00:00Z"}`
		switch got["path"] {
		case "/workspace/outdir":
			fmt.Fprintf(w, `{"created":true,"entry":%s}`, entry)
		case "/workspace/existing":
			fmt.Fprintf(w, `{"entry":%s}`, entry)
		default:
			fmt.Fprint(w, `{}`)
		}
	})

	got, _, err := client.Sandboxes.MakeDirectory(ctx, "sb-1", "/workspace/outdir")
	if err != nil || !got.Created || got.Entry.Type != SandboxFileEntryTypeDirectory || got.Entry.Mode != 493 {
		t.Fatalf("MakeDirectory returned %+v, %v", got, err)
	}
	if got, _, err := client.Sandboxes.MakeDirectory(ctx, "sb-1", "/workspace/existing"); err != nil || got.Created {
		t.Fatalf("an existing directory returned %+v, %v", got, err)
	}
	if _, _, err := client.Sandboxes.MakeDirectory(ctx, "sb-1", "/workspace/empty"); err == nil {
		t.Error("a response without an entry returned no error")
	}
}

func TestSandboxes_MoveAndRemoveFile(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/files/move", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got map[string]string
		_ = json.NewDecoder(r.Body).Decode(&got)
		if want := map[string]string{"source": "/workspace/old.txt", "destination": "/workspace/hello.txt"}; !reflect.DeepEqual(got, want) {
			t.Errorf("request body = %+v, expected %+v", got, want)
		}
		fmt.Fprintf(w, `{"entry":%s}`, sandboxFileEntryJSON)
	})
	mux.HandleFunc("/v2/sandboxes/sb-1/files/remove", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got map[string]string
		_ = json.NewDecoder(r.Body).Decode(&got)
		if got["path"] != "/workspace/old.txt" {
			t.Errorf("request body = %+v", got)
		}
		fmt.Fprint(w, `{}`)
	})

	got, _, err := client.Sandboxes.MoveFile(ctx, "sb-1", "/workspace/old.txt", "/workspace/hello.txt")
	if err != nil || got.Path != "/workspace/hello.txt" {
		t.Fatalf("MoveFile returned %+v, %v", got, err)
	}
	if _, _, err := client.Sandboxes.MoveFile(ctx, "sb-1", "/a", ""); err == nil {
		t.Error("an empty destination returned no error")
	}
	if _, err := client.Sandboxes.RemoveFile(ctx, "sb-1", "/workspace/old.txt"); err != nil {
		t.Fatalf("RemoveFile returned %v", err)
	}
	if _, err := client.Sandboxes.RemoveFile(ctx, "sb-1", ""); err == nil {
		t.Error("an empty path returned no error")
	}
}
