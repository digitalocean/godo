// Command sandbox creates a DigitalOcean Sandbox, pauses and resumes it, lists
// the running sandboxes and deletes it.
//
//	DIGITALOCEAN_ACCESS_TOKEN=... go run ./examples/sandbox
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/sandbox"
)

func main() {
	ctx := context.Background()

	c, err := sandbox.NewClient(nil) // reads DIGITALOCEAN_ACCESS_TOKEN
	if err != nil {
		log.Fatal(err)
	}

	sb, err := c.Create(ctx, nil) // coding-base, mars-2vcpu-4gb
	if err != nil {
		log.Fatal(err)
	}
	defer sb.Kill(ctx)
	fmt.Println("created", sb.ID(), sb.State())

	if err := sb.Pause(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("paused", sb.State())
	if err := sb.Resume(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("resumed", sb.State())

	for s, err := range c.List(&sandbox.ListParams{State: godo.SandboxStateRunning}).All(ctx) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("running:", s.ID(), s.Template())
	}
}
