package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/digitalocean/godo"
)

// download-workspace streams a file (or tar archive) out of a session's sandbox
// workspace and writes it to disk.
//
// The download is chunked and ends with a fixed 73-byte integrity footer
// (DOWSSHA1 + SHA-256 hex + newline). The godo client strips that footer and
// verifies the payload checksum for you: you MUST read the body to EOF
// (io.Copy below) and then Close it. A missing, invalid, or mismatched footer
// surfaces as an error — when that happens the partial output must be discarded.
//
// Required env:
//   - DIGITALOCEAN_TOKEN
//   - HOSTED_AGENT_SESSION_ID
//   - WORKSPACE_PATH          source path inside /workspace
//   - LOCAL_FILE              local destination path
//
// Optional env:
//   - AS_ARCHIVE=true         tar-stream the directory at WORKSPACE_PATH
func main() {
	sessionID := mustEnv("HOSTED_AGENT_SESSION_ID")
	workspacePath := mustEnv("WORKSPACE_PATH")
	localFile := mustEnv("LOCAL_FILE")

	client := mustClient()
	ctx := context.Background()

	dl, resp, err := client.HostedAgents.DownloadWorkspace(ctx, sessionID, &godo.HostedAgentWorkspaceDownloadRequest{
		Path:      workspacePath,
		AsArchive: os.Getenv("AS_ARCHIVE") == "true",
	})
	if err != nil {
		die(err)
	}
	defer dl.Body.Close()

	fmt.Printf("HTTP %d\n", resp.StatusCode)
	fmt.Printf("is_archive: %t\n", dl.IsArchive)
	if dl.SizeBytes > 0 {
		fmt.Printf("size hint:  %d bytes\n", dl.SizeBytes)
	}

	out, err := os.Create(localFile)
	if err != nil {
		die(err)
	}

	// Read to EOF. The final Read verifies the integrity footer, so an error
	// here means the transfer was truncated or corrupted.
	n, copyErr := io.Copy(out, dl.Body)
	closeErr := out.Close()

	if copyErr != nil {
		os.Remove(localFile) // discard corrupted output
		die(fmt.Errorf("download failed after %d bytes (output discarded): %w", n, copyErr))
	}
	if closeErr != nil {
		die(closeErr)
	}

	fmt.Printf("wrote:      %d bytes to %s\n", n, localFile)
	fmt.Println("checksum:   verified")
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
