package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/digitalocean/godo"
)

// upload-workspace streams a local file into a session's sandbox workspace.
//
// Required env:
//   - DIGITALOCEAN_TOKEN
//   - HOSTED_AGENT_SESSION_ID
//   - LOCAL_FILE              path to the file to upload
//   - WORKSPACE_PATH          destination path inside /workspace
//
// Optional env:
//   - IS_ARCHIVE=true         treat LOCAL_FILE as a tar to extract at WORKSPACE_PATH
//   - SEND_CHECKSUM=true      compute and send the X-Content-Sha256 header
func main() {
	sessionID := mustEnv("HOSTED_AGENT_SESSION_ID")
	localFile := mustEnv("LOCAL_FILE")
	workspacePath := mustEnv("WORKSPACE_PATH")

	f, err := os.Open(localFile)
	if err != nil {
		die(err)
	}
	defer f.Close()

	req := &godo.HostedAgentWorkspaceUploadRequest{
		Path:      workspacePath,
		IsArchive: os.Getenv("IS_ARCHIVE") == "true",
		Body:      f,
	}

	// Optionally hash the file up front and forward the digest so the guest can
	// verify the upload. This buffers the file once to compute the digest, then
	// rewinds before streaming it.
	if os.Getenv("SEND_CHECKSUM") == "true" {
		sum, err := fileSHA256(localFile)
		if err != nil {
			die(err)
		}
		req.ContentSHA256 = sum
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			die(err)
		}
		fmt.Printf("X-Content-Sha256: %s\n", sum)
	}

	client := mustClient()
	ctx := context.Background()

	out, resp, err := client.HostedAgents.UploadWorkspace(ctx, sessionID, req)
	if err != nil {
		die(err)
	}

	fmt.Printf("HTTP %d\n", resp.StatusCode)
	fmt.Printf("path:          %s\n", out.Path)
	fmt.Printf("bytes_written: %d\n", out.BytesWritten)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "%s is required\n", key)
		os.Exit(2)
	}
	return v
}

func mustClient() *godo.Client {
	token := mustEnv("DIGITALOCEAN_TOKEN")
	client := godo.NewFromToken(token)
	if baseURL := os.Getenv("DIGITALOCEAN_API_URL"); baseURL != "" {
		u, err := url.Parse(baseURL)
		if err != nil {
			panic(err)
		}
		client.BaseURL = u
	}
	return client
}

func die(err error) {
	var apiErr *godo.ErrorResponse
	if errors.As(err, &apiErr) {
		fmt.Fprintf(os.Stderr, "API error (HTTP %d): %s\n", apiErr.Response.StatusCode, apiErr.Message)
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(1)
}
