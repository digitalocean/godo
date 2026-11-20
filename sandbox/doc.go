// Package sandbox is the DigitalOcean Sandbox SDK: a small client over the
// /v2/sandboxes API with defaults, waiting and paging.
//
//	c, err := sandbox.NewClient(nil) // reads DIGITALOCEAN_ACCESS_TOKEN
//	if err != nil { ... }
//
//	sb, err := c.Create(ctx, nil) // coding-base, mars-2vcpu-4gb
//	if err != nil { ... }
//	defer sb.Kill(ctx)
//
//	_ = sb.Pause(ctx)
//	_ = sb.Resume(ctx)
//
// Every call goes through godo's Sandboxes service, so auth, retries and
// User-Agent match the rest of godo. Use that service
// (godo.Client.Sandboxes) directly when you want the raw API.
package sandbox
