package sandbox_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/digitalocean/godo/sandbox"
)

const fileEntryJSON = `{"name":"hello.txt","type":"FILE_ENTRY_TYPE_FILE","path":"/workspace/hello.txt","size":"12","mode":420,"permissions":"-rw-r--r--","owner":"user","group":"user","modified_time":"2026-03-25T12:00:00Z"}`

func sha(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestFilesWriteSendsTheContentAndItsChecksum(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	data := []byte("hello world\n")
	a.handlers["POST /v2/sandboxes/sb-1/files/upload"] = func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query(); q.Get("path") != "/workspace/hello.txt" || q.Has("is_archive") {
			t.Errorf("query = %v", q)
		}
		if got := r.Header.Get("X-Content-Sha256"); got != sha(data) {
			t.Errorf("X-Content-Sha256 = %q", got)
		}
		fmt.Fprint(w, `{"path":"/workspace/hello.txt","bytes_written":"12"}`)
	}

	info, err := sb.Files.Write(context.Background(), "/workspace/hello.txt", data)
	if err != nil || info.Path != "/workspace/hello.txt" || info.BytesWritten != 12 {
		t.Fatalf("Write = %+v, %v", info, err)
	}
	if got := string(a.raw[len(a.raw)-1]); got != string(data) {
		t.Errorf("body = %q", got)
	}
}

func TestFilesWriteArchiveExtractsIntoADirectory(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers["POST /v2/sandboxes/sb-1/files/upload"] = func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query(); q.Get("path") != "/workspace/out" || q.Get("is_archive") != "true" {
			t.Errorf("query = %v", q)
		}
		fmt.Fprint(w, `{"path":"/workspace/out","bytes_written":"3"}`)
	}
	if info, err := sb.Files.WriteArchive(context.Background(), "/workspace/out", []byte("tar")); err != nil || info.BytesWritten != 3 {
		t.Fatalf("WriteArchive = %+v, %v", info, err)
	}
}

func TestFilesReadVerifiesAndStripsTheFooter(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	ctx := context.Background()
	payload := []byte("hello world\n")
	a.handlers["GET /v2/sandboxes/sb-1/files/download"] = func(w http.ResponseWriter, r *http.Request) {
		body := payload
		if r.URL.Query().Get("as_archive") == "true" {
			body = []byte("tarball")
		}
		if r.URL.Query().Get("path") == "/bad" {
			fmt.Fprint(w, "tampered"+"DOWSSHA1"+sha(payload)+"\n")
			return
		}
		fmt.Fprint(w, string(body)+"DOWSSHA1"+sha(body)+"\n")
	}

	if got, err := sb.Files.Read(ctx, "/workspace/hello.txt"); err != nil || string(got) != string(payload) {
		t.Fatalf("Read = %q, %v", got, err)
	}
	if got, err := sb.Files.ReadArchive(ctx, "/workspace"); err != nil || string(got) != "tarball" {
		t.Fatalf("ReadArchive = %q, %v", got, err)
	}
	if _, err := sb.Files.Read(ctx, "/bad"); err == nil {
		t.Error("a checksum mismatch returned no error")
	}
}

func TestFilesListAndGetInfoNormalizeEntries(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	ctx := context.Background()
	a.handlers["GET /v2/sandboxes/sb-1/files/list"] = func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query(); q.Get("path") != "/workspace" || q.Get("depth") != "2" {
			t.Errorf("query = %v", q)
		}
		fmt.Fprintf(w, `{"entries":[%s,{"name":"d","type":"FILE_ENTRY_TYPE_DIRECTORY","path":"/workspace/d"},{"name":"l","type":"FILE_ENTRY_TYPE_SYMLINK","path":"/workspace/l","symlink_target":"d"},{"name":"x"}]}`, fileEntryJSON)
	}
	a.handlers["GET /v2/sandboxes/sb-1/files/info"] = func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("path") == "/missing" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"id":"not_found","message":"no such file"}`)
			return
		}
		fmt.Fprintf(w, `{"entry":%s}`, fileEntryJSON)
	}

	entries, err := sb.Files.List(ctx, "/workspace", &sandbox.ListFilesOptions{Depth: 2})
	if err != nil || len(entries) != 4 {
		t.Fatalf("List = %+v, %v", entries, err)
	}
	want := sandbox.FileEntry{
		Name: "hello.txt", Type: sandbox.FileTypeFile, Path: "/workspace/hello.txt", Size: 12, Mode: 0o644,
		Permissions: "-rw-r--r--", Owner: "user", Group: "user",
		ModifiedTime: time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC),
	}
	if entries[0] != want {
		t.Errorf("entries[0] = %+v", entries[0])
	}
	if !entries[1].IsDir() || entries[2].Type != sandbox.FileTypeSymlink || entries[2].SymlinkTarget != "d" || entries[3].Type != sandbox.FileTypeUnspecified {
		t.Errorf("entries = %+v", entries)
	}

	info, err := sb.Files.GetInfo(ctx, "/workspace/hello.txt")
	if err != nil || *info != want {
		t.Fatalf("GetInfo = %+v, %v", info, err)
	}
	if _, err := sb.Files.GetInfo(ctx, "/missing"); !sandbox.IsNotFound(err) {
		t.Errorf("a missing path returned %v", err)
	}
}

func TestFilesExistsMakeDirMoveAndRemove(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	ctx := context.Background()
	a.handlers["GET /v2/sandboxes/sb-1/files/exists"] = func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("path") == "/workspace/hello.txt" {
			fmt.Fprint(w, `{"exists":true}`)
			return
		}
		fmt.Fprint(w, `{}`)
	}
	a.handlers["POST /v2/sandboxes/sb-1/files/make-directory"] = reply(200, `{"created":true,"entry":{"name":"out","type":"FILE_ENTRY_TYPE_DIRECTORY","path":"/workspace/out","mode":493}}`)
	a.handlers["POST /v2/sandboxes/sb-1/files/move"] = reply(200, `{"entry":`+fileEntryJSON+`}`)
	a.handlers["POST /v2/sandboxes/sb-1/files/remove"] = reply(200, `{}`)

	if ok, err := sb.Files.Exists(ctx, "/workspace/hello.txt"); err != nil || !ok {
		t.Fatalf("Exists = %v, %v", ok, err)
	}
	if ok, err := sb.Files.Exists(ctx, "/workspace/nope"); err != nil || ok {
		t.Fatalf("Exists({}) = %v, %v", ok, err)
	}

	created, dir, err := sb.Files.MakeDir(ctx, "/workspace/out")
	if err != nil || !created || !dir.IsDir() || dir.Mode != 0o755 {
		t.Fatalf("MakeDir = %v, %+v, %v", created, dir, err)
	}
	if body := a.bodies[len(a.bodies)-1]; body["path"] != "/workspace/out" {
		t.Errorf("make-directory body = %+v", body)
	}

	moved, err := sb.Files.Move(ctx, "/workspace/old.txt", "/workspace/hello.txt")
	if err != nil || moved.Path != "/workspace/hello.txt" {
		t.Fatalf("Move = %+v, %v", moved, err)
	}
	if body := a.bodies[len(a.bodies)-1]; body["source"] != "/workspace/old.txt" || body["destination"] != "/workspace/hello.txt" {
		t.Errorf("move body = %+v", body)
	}

	if err := sb.Files.Remove(ctx, "/workspace/old.txt"); err != nil {
		t.Fatal(err)
	}
	if body := a.bodies[len(a.bodies)-1]; body["path"] != "/workspace/old.txt" {
		t.Errorf("remove body = %+v", body)
	}
}

func TestFilesRejectEmptyPathsBeforeSending(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	ctx := context.Background()
	calls := len(a.calls)

	checks := map[string]error{}
	_, checks["Write"] = sb.Files.Write(ctx, "", []byte("x"))
	_, checks["WriteArchive"] = sb.Files.WriteArchive(ctx, " ", nil)
	_, checks["Read"] = sb.Files.Read(ctx, "")
	_, checks["ReadArchive"] = sb.Files.ReadArchive(ctx, "")
	_, checks["List"] = sb.Files.List(ctx, "", nil)
	_, checks["List depth"] = sb.Files.List(ctx, "/workspace", &sandbox.ListFilesOptions{Depth: -1})
	_, checks["GetInfo"] = sb.Files.GetInfo(ctx, "")
	_, checks["Exists"] = sb.Files.Exists(ctx, "")
	_, _, checks["MakeDir"] = sb.Files.MakeDir(ctx, "")
	_, checks["Move source"] = sb.Files.Move(ctx, "", "/b")
	_, checks["Move destination"] = sb.Files.Move(ctx, "/a", "")
	checks["Remove"] = sb.Files.Remove(ctx, "")
	for name, err := range checks {
		if err == nil {
			t.Errorf("%s returned no error", name)
		}
	}
	if len(a.calls) != calls {
		t.Errorf("a request was sent for an empty path: %v", a.calls[calls:])
	}
}
