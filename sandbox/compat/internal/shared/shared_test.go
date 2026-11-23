package shared

import (
	"io"
	"testing"
	"time"

	"github.com/digitalocean/godo"
)

func TestSizeFor(t *testing.T) {
	for _, tc := range []struct {
		cpu  float64
		mem  int
		want string
	}{
		{0, 0, ""}, {1, 512, "mars-1vcpu-1gb"}, {1.5, 0, "mars-2vcpu-2gb"}, {2, 4096, "mars-2vcpu-4gb"}, {0, 6000, "mars-4vcpu-8gb"}, {16, 32768, "mars-16vcpu-32gb"},
	} {
		got, err := SizeFor(tc.cpu, tc.mem)
		if err != nil || got != tc.want {
			t.Errorf("SizeFor(%v, %d) = %q, %v; want %q", tc.cpu, tc.mem, got, err, tc.want)
		}
	}
	if _, err := SizeFor(64, 0); err == nil {
		t.Error("an oversized request returned no error")
	}
	if s, ok := SizeInfo("mars-2vcpu-4gb"); !ok || s.VCPUs != 2 || s.MemoryMiB != 4096 {
		t.Errorf("SizeInfo = %+v, %v", s, ok)
	}
}

func TestTemplateNameIsStable(t *testing.T) {
	a, b := TemplateName("daytona", "python:3.12"), TemplateName("daytona", "python:3.12")
	if a != b || a == TemplateName("daytona", "python:3.13") || len(a) != len("daytona-")+12 {
		t.Errorf("names: %q %q", a, b)
	}
}

func TestImageRef(t *testing.T) {
	for _, tc := range []struct {
		src  *godo.SandboxImageSource
		want string
	}{
		{nil, ""},
		{&godo.SandboxImageSource{Registry: "docker.io", Repository: "library/python", Tag: "3.12"}, "docker.io/library/python:3.12"},
		{&godo.SandboxImageSource{Repository: "python", Digest: "sha256:abc"}, "python@sha256:abc"},
	} {
		if got := ImageRef(tc.src); got != tc.want {
			t.Errorf("ImageRef(%+v) = %q, want %q", tc.src, got, tc.want)
		}
	}
}

func TestCeilSeconds(t *testing.T) {
	for d, want := range map[time.Duration]int64{0: 0, -1: 0, time.Millisecond: 1, time.Second: 1, 1500 * time.Millisecond: 2} {
		if got := CeilSeconds(d); got != want {
			t.Errorf("CeilSeconds(%v) = %d, want %d", d, got, want)
		}
	}
}

func TestPipeBuffersAndEnds(t *testing.T) {
	p := NewPipe()
	if _, err := p.Write([]byte("hello ")); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Write([]byte("world")); err != nil {
		t.Fatal(err)
	}
	p.CloseWrite()
	got, err := io.ReadAll(p)
	if err != nil || string(got) != "hello world" {
		t.Errorf("ReadAll = %q, %v", got, err)
	}
	if _, err := p.Write([]byte("x")); err == nil {
		t.Error("a write after CloseWrite succeeded")
	}

	q := NewPipe()
	done := make(chan struct{})
	go func() { _, _ = io.ReadAll(q); close(done) }()
	_ = q.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Close did not unblock a reader")
	}
	if n, err := q.Write([]byte("dropped")); err != nil || n != 7 {
		t.Errorf("Write after the reader closed = %d, %v", n, err)
	}
}
