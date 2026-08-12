package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

func main() {
	client := mustClient()
	ctx := context.Background()

	sessionID := os.Getenv("HOSTED_AGENT_SESSION_ID")

	if sessionID == "" {
		ok("CreateSession", func() error {
			s, _, err := client.HostedAgents.CreateSession(ctx, &godo.HostedAgentSessionCreateRequest{
				AgentKind: godo.HostedAgentKindClaudeCode,
				RepoHint:  envOr("REPO_HINT", "digitalocean/godo"),
			})
			if err != nil {
				return err
			}
			if s == nil || s.SessionID == "" {
				return errors.New("empty session in response")
			}
			sessionID = s.SessionID
			fmt.Printf("       → session_id=%s status=%s\n", s.SessionID, s.Status)
			return nil
		})
	} else {
		fmt.Printf("SKIP CreateSession (using HOSTED_AGENT_SESSION_ID=%s)\n", sessionID)
	}

	if sessionID == "" {
		os.Exit(1)
	}

	ok("GetSession", func() error {
		s, _, err := client.HostedAgents.GetSession(ctx, sessionID)
		if err != nil {
			return err
		}
		if s.SessionID != sessionID {
			return fmt.Errorf("got session_id %q, want %q", s.SessionID, sessionID)
		}
		fmt.Printf("       → status=%s sandbox=%s\n", s.Status, s.SandboxID)
		return nil
	})

	ok("ListSessions", func() error {
		out, _, err := client.HostedAgents.ListSessions(ctx, &godo.HostedAgentSessionListOptions{
			PageSize: 50,
		})
		if err != nil {
			return err
		}
		found := false
		for _, s := range out.Sessions {
			if s.SessionID == sessionID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("session %q not in list (%d returned)", sessionID, len(out.Sessions))
		}
		fmt.Printf("       → listed %d session(s), includes target\n", len(out.Sessions))
		return nil
	})

	ok("StreamSession", func() error {
		streamCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		stream, _, err := client.HostedAgents.StreamSession(streamCtx, sessionID, nil)
		if err != nil {
			return err
		}
		defer stream.Close()

		var n int
		for stream.Next() {
			ev := stream.Current()
			fmt.Printf("       → event kind=%s id=%s\n", ev.Kind, ev.EventID)
			n++
			if n >= 5 {
				break
			}
		}
		if err := stream.Err(); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		fmt.Printf("       → received %d event(s)\n", n)
		return nil
	})

	ok("SendInput", func() error {
		out, _, err := client.HostedAgents.SendInput(ctx, sessionID, &godo.HostedAgentSendInputRequest{
			Text: envOr("INPUT_TEXT", "Say hello in one sentence."),
		})
		if err != nil {
			return err
		}
		if out.RunID == "" {
			return errors.New("empty run_id in response")
		}
		fmt.Printf("       → run_id=%s\n", out.RunID)
		return nil
	})

	skipOrOK("ExecInSandbox", func() (skip bool, err error) {
		_, resp, err := client.HostedAgents.ExecInSandbox(ctx, sessionID, &godo.HostedAgentSandboxExecRequest{
			Argv: []string{"echo", "hello"},
		})
		if err != nil {
			var apiErr *godo.ErrorResponse
			if errors.As(err, &apiErr) && apiErr.Response.StatusCode == 501 {
				return true, nil
			}
			return false, err
		}
		_ = resp
		return false, nil
	})

	if os.Getenv("HITL_REQUEST_ID") != "" {
		ok("ResolveHITL", func() error {
			_, err := client.HostedAgents.ResolveHITL(ctx, sessionID, os.Getenv("HITL_REQUEST_ID"),
				&godo.HostedAgentResolveHITLRequest{
					Outcome: godo.HostedAgentHITLOutcomeApprove,
				})
			return err
		})
	} else {
		fmt.Println("SKIP ResolveHITL (set HITL_REQUEST_ID when a HITL prompt is pending)")
	}

	if os.Getenv("DESTROY_SESSION") == "true" {
		ok("DestroySession", func() error {
			resp, err := client.HostedAgents.DestroySession(ctx, sessionID)
			if err != nil {
				return err
			}
			if resp.StatusCode != 204 {
				return fmt.Errorf("expected HTTP 204, got %d", resp.StatusCode)
			}
			return nil
		})
	} else {
		fmt.Printf("SKIP DestroySession (session %s left running; set DESTROY_SESSION=true to delete)\n", sessionID)
	}
}

func ok(name string, fn func() error) {
	if err := fn(); err != nil {
		fmt.Printf("FAIL %s: %v\n", name, err)
		os.Exit(1)
	}
	fmt.Printf("OK   %s\n", name)
}

func skipOrOK(name string, fn func() (skip bool, err error)) {
	skip, err := fn()
	if err != nil {
		fmt.Printf("FAIL %s: %v\n", name, err)
		os.Exit(1)
	}
	if skip {
		fmt.Printf("SKIP %s (not implemented on this server)\n", name)
		return
	}
	fmt.Printf("OK   %s\n", name)
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

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
