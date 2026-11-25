// Package sandbox is the DigitalOcean Sandbox SDK: a small client over the
// /v2/sandboxes API with defaults, waiting and paging, commands, exec and file
// transfer.
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
//	_, err = sb.Files.Write(ctx, "/workspace/hello.txt", []byte("hello\n"))
//	data, err := sb.Files.Read(ctx, "/workspace/hello.txt")
//
//	res, err := sb.Exec.RunShell(ctx, "cat /workspace/hello.txt", nil)
//	fmt.Print(res.Stdout, res.ExitCode, res.OK)
//
// A non-zero exit is not an error unless ExecOptions.Check is set: then Run and
// RunShell return the result and a *CommandError. ExecOptions.Timeout sets the
// timeout as a time.Duration. Go keeps sb.Exec.Run because Go cannot call a
// struct field.
//
// Every call goes through godo's Sandboxes service, so auth, retries and
// User-Agent match the rest of godo. Use that service
// (godo.Client.Sandboxes) directly when you want the raw API.
package sandbox
