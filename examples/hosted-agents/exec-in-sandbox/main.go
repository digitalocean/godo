package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/digitalocean/godo"
)

func main() {
	sessionID := os.Getenv("HOSTED_AGENT_SESSION_ID")
	if sessionID == "" {
		fmt.Fprintln(os.Stderr, "HOSTED_AGENT_SESSION_ID is required")
		os.Exit(2)
	}

	raw := envOr("EXEC_COMMAND", "echo,hello from sandbox")
	argv := strings.Split(raw, ",")
	for i := range argv {
		argv[i] = strings.TrimSpace(argv[i])
	}

	client := mustClient()
	ctx := context.Background()

	out, resp, err := client.HostedAgents.ExecInSandbox(ctx, sessionID, &godo.HostedAgentSandboxExecRequest{
		Argv:           argv,
		Workdir:        os.Getenv("EXEC_WORKDIR"),
		TimeoutSeconds: 30,
	})
	if err != nil {
		var apiErr *godo.ErrorResponse
		if errors.As(err, &apiErr) && apiErr.Response.StatusCode == 501 {
			fmt.Fprintf(os.Stderr, "HTTP 501 — ExecInSandbox is not implemented on this server yet\n")
			os.Exit(1)
		}
		die(err)
	}

	fmt.Printf("HTTP %d\n", resp.StatusCode)
	fmt.Printf("exit_code: %d\n", out.ExitCode)
	if out.Stdout != "" {
		fmt.Printf("stdout:\n%s\n", out.Stdout)
	}
	if out.Stderr != "" {
		fmt.Printf("stderr:\n%s\n", out.Stderr)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustClient() *godo.Client {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "DIGITALOCEAN_TOKEN is required")
		os.Exit(2)
	}
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
