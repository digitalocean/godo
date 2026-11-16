// Smoke the ADS client against a live API.
//
//	DIGITALOCEAN_TOKEN=... go run ./examples/ads/smoke
//
// Optional:
//
//	DIGITALOCEAN_API_URL   API base override, including trailing slash (staging)
//	ADS_REGION             region slug (default tor1)
//	ADS_WORKSPACE_NAME     workspace name (default ads-smoke)
//	ADS_QUERY_SIZE         query_instance_size (default small)
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/digitalocean/godo"
)

// cleanup, when set, runs if a later step fails so we don't leak a workspace.
var cleanup func()

func main() {
	client := mustClient()
	ctx := context.Background()

	region := envOr("ADS_REGION", "tor1")
	name := envOr("ADS_WORKSPACE_NAME", fmt.Sprintf("ads-smoke-%d", os.Getpid()))
	size := envOr("ADS_QUERY_SIZE", "small")

	fmt.Printf("base_url=%s region=%s name=%s size=%s\n", client.BaseURL, region, name, size)

	// ListPlans is skipped: GET /v2/data-cloud-platform/plans is not routed on Edge yet.

	var ws *godo.ADSWorkspace
	ok("CreateWorkspace", func() error {
		created, resp, err := client.ADS.CreateWorkspace(ctx, region, &godo.ADSWorkspaceCreateRequest{
			Name:              name,
			QueryInstanceSize: size,
		})
		if err != nil {
			return fmt.Errorf("%w (status=%d req=%s)", err, status(resp), reqURL(resp))
		}
		ws = created
		cleanup = func() {
			fmt.Printf("       → cleanup: deleting uuid=%s\n", ws.UUID)
			if _, err := client.ADS.DeleteWorkspace(ctx, region, ws.UUID); err != nil {
				fmt.Printf("       → cleanup failed (delete %s manually): %v\n", ws.UUID, err)
			}
		}
		fmt.Printf("       → uuid=%s status=%s region=%s\n", ws.UUID, ws.Status, ws.Region)
		return nil
	})

	ok("ListWorkspaces", func() error {
		list, resp, err := client.ADS.ListWorkspaces(ctx, region)
		if err != nil {
			return fmt.Errorf("%w (status=%d req=%s)", err, status(resp), reqURL(resp))
		}
		fmt.Printf("       → %d workspace(s)\n", len(list.Workspaces))
		return nil
	})

	ok("GetWorkspace", func() error {
		got, resp, err := client.ADS.GetWorkspace(ctx, region, ws.UUID)
		if err != nil {
			return fmt.Errorf("%w (status=%d req=%s)", err, status(resp), reqURL(resp))
		}
		fmt.Printf("       → uuid=%s status=%s name=%s\n", got.UUID, got.Status, got.Name)
		return nil
	})

	ok("DeleteWorkspace", func() error {
		resp, err := client.ADS.DeleteWorkspace(ctx, region, ws.UUID)
		if err != nil {
			return fmt.Errorf("%w (status=%d req=%s)", err, status(resp), reqURL(resp))
		}
		cleanup = nil // already deleted
		fmt.Printf("       → deleted uuid=%s\n", ws.UUID)
		return nil
	})
}

func ok(name string, fn func() error) {
	if err := fn(); err != nil {
		fmt.Printf("FAIL %s: %v\n", name, err)
		if cleanup != nil {
			cleanup()
		}
		os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "invalid DIGITALOCEAN_API_URL: %v\n", err)
			os.Exit(2)
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

func status(resp *godo.Response) int {
	if resp == nil || resp.Response == nil {
		return 0
	}
	return resp.StatusCode
}

func reqURL(resp *godo.Response) string {
	if resp == nil || resp.Response == nil || resp.Request == nil {
		return ""
	}
	return resp.Request.Method + " " + resp.Request.URL.String()
}