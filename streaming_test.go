package godo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSSEReader_SingleEvent(t *testing.T) {
	r := NewSSEReader(strings.NewReader("data: hello\n\n"))
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("Next: unexpected error %v", err)
	}
	if ev.Event != "message" {
		t.Errorf("Event = %q, want %q", ev.Event, "message")
	}
	if string(ev.Data) != "hello" {
		t.Errorf("Data = %q, want %q", ev.Data, "hello")
	}
	if _, err := r.Next(); err != io.EOF {
		t.Errorf("second Next: err = %v, want io.EOF", err)
	}
}

func TestSSEReader_MultiDataLinesJoined(t *testing.T) {
	r := NewSSEReader(strings.NewReader("data: line1\ndata: line2\ndata: line3\n\n"))
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("Next: unexpected error %v", err)
	}
	if string(ev.Data) != "line1\nline2\nline3" {
		t.Errorf("Data = %q, want %q", ev.Data, "line1\nline2\nline3")
	}
}

func TestSSEReader_CommentsSkipped(t *testing.T) {
	in := ": keep-alive\n: another comment\ndata: payload\n\n"
	r := NewSSEReader(strings.NewReader(in))
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("Next: unexpected error %v", err)
	}
	if string(ev.Data) != "payload" {
		t.Errorf("Data = %q, want %q", ev.Data, "payload")
	}
}

func TestSSEReader_EventIDAndRetry(t *testing.T) {
	in := "event: chunk\nid: 42\nretry: 1500\ndata: hi\n\ndata: again\n\n"
	r := NewSSEReader(strings.NewReader(in))

	ev, err := r.Next()
	if err != nil {
		t.Fatalf("Next #1: %v", err)
	}
	if ev.Event != "chunk" {
		t.Errorf("Event = %q, want chunk", ev.Event)
	}
	if ev.ID != "42" {
		t.Errorf("ID = %q, want 42", ev.ID)
	}
	if ev.Retry != 1500 {
		t.Errorf("Retry = %d, want 1500", ev.Retry)
	}
	if string(ev.Data) != "hi" {
		t.Errorf("Data = %q, want hi", ev.Data)
	}

	ev, err = r.Next()
	if err != nil {
		t.Fatalf("Next #2: %v", err)
	}
	if ev.Event != "message" {
		t.Errorf("Event = %q, want message (default)", ev.Event)
	}
	if ev.ID != "42" {
		t.Errorf("ID persistence: got %q, want 42", ev.ID)
	}
	if string(ev.Data) != "again" {
		t.Errorf("Data = %q, want again", ev.Data)
	}
}

func TestSSEReader_CRLFTerminators(t *testing.T) {
	r := NewSSEReader(strings.NewReader("data: a\r\ndata: b\r\n\r\n"))
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if string(ev.Data) != "a\nb" {
		t.Errorf("Data = %q, want %q", ev.Data, "a\nb")
	}
}

func TestSSEReader_LargePayload(t *testing.T) {
	const size = 512 * 1024
	payload := bytes.Repeat([]byte("x"), size)
	in := bytes.NewBuffer(nil)
	in.WriteString("data: ")
	in.Write(payload)
	in.WriteString("\n\n")

	r := NewSSEReader(in)
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if len(ev.Data) != size {
		t.Fatalf("Data length = %d, want %d", len(ev.Data), size)
	}
	if !bytes.Equal(ev.Data, payload) {
		t.Fatalf("Data payload mismatch")
	}
}

func TestSSEReader_TrailingEventWithoutBlankLine(t *testing.T) {
	r := NewSSEReader(strings.NewReader("data: tail"))
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if string(ev.Data) != "tail" {
		t.Errorf("Data = %q, want tail", ev.Data)
	}
	if _, err := r.Next(); err != io.EOF {
		t.Errorf("subsequent Next: err = %v, want io.EOF", err)
	}
}

func TestSSEReader_NoLeadingSpaceValue(t *testing.T) {
	r := NewSSEReader(strings.NewReader("data:no-space\ndata:  two-spaces\n\n"))
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if string(ev.Data) != "no-space\n two-spaces" {
		t.Errorf("Data = %q, want %q", ev.Data, "no-space\n two-spaces")
	}
}

func TestDoStream_HappyPath(t *testing.T) {
	setup()
	defer teardown()

	chunks := []string{
		"data: {\"i\":1}\n\n",
		": keep-alive\n\n",
		"data: {\"i\":2}\n\n",
		"data: [DONE]\n\n",
	}

	mux.HandleFunc("/v1/stream", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter is not a Flusher")
		}
		for _, c := range chunks {
			if _, err := io.WriteString(w, c); err != nil {
				return
			}
			flusher.Flush()
		}
	})

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/v1/stream", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := client.DoStream(context.Background(), req)
	if err != nil {
		t.Fatalf("DoStream: %v", err)
	}
	defer resp.Body.Close()

	r := NewSSEReader(resp.Body)
	var got []string
	for {
		ev, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("SSE Next: %v", err)
		}
		got = append(got, string(ev.Data))
	}

	want := []string{`{"i":1}`, `{"i":2}`, "[DONE]"}
	if !equalStrings(got, want) {
		t.Errorf("events = %v, want %v", got, want)
	}
}

func TestDoStream_ProgressiveDelivery(t *testing.T) {
	setup()
	defer teardown()

	release := make(chan struct{})
	done := make(chan struct{})

	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		io.WriteString(w, "data: first\n\n")
		flusher.Flush()
		<-release
		io.WriteString(w, "data: second\n\n")
		flusher.Flush()
		close(done)
	})

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/stream", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := client.DoStream(context.Background(), req)
	if err != nil {
		t.Fatalf("DoStream: %v", err)
	}
	defer resp.Body.Close()

	r := NewSSEReader(resp.Body)

	ev, err := r.Next()
	if err != nil {
		t.Fatalf("first Next: %v", err)
	}
	if string(ev.Data) != "first" {
		t.Fatalf("first event = %q, want first", ev.Data)
	}

	close(release)
	ev, err = r.Next()
	if err != nil {
		t.Fatalf("second Next: %v", err)
	}
	if string(ev.Data) != "second" {
		t.Fatalf("second event = %q, want second", ev.Data)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("server handler did not complete")
	}
}

func TestDoStream_ErrorResponseClosesBody(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/boom", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"id":"unauthorized","message":"no token"}`)
	})

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/boom", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := client.DoStream(context.Background(), req)
	if err == nil {
		t.Fatal("DoStream: expected error, got nil")
	}
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *ErrorResponse", err)
	}
	if apiErr.Message != "no token" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "no token")
	}
	if resp == nil {
		t.Fatal("DoStream: response was nil on HTTP error")
	}
	if err := resp.Body.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}

func TestDoStream_ContextCancellation(t *testing.T) {
	setup()
	defer teardown()

	serverReady := make(chan struct{})
	mux.HandleFunc("/hang", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		close(serverReady)
		<-r.Context().Done()
	})

	ctx, cancel := context.WithCancel(context.Background())
	req, err := client.NewRequest(ctx, http.MethodGet, "/hang", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := client.DoStream(ctx, req)
	if err != nil {
		t.Fatalf("DoStream: %v", err)
	}
	defer resp.Body.Close()

	<-serverReady

	readErr := make(chan error, 1)
	go func() {
		_, err := NewSSEReader(resp.Body).Next()
		readErr <- err
	}()

	cancel()

	select {
	case err := <-readErr:
		if err == nil {
			t.Fatal("Next returned nil error after cancellation")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Next did not return after context cancellation")
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
