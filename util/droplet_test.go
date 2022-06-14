package util

import (
	"context"
	"fmt"
	"log"

	"github.com/digitalocean/godo"
)

func ExampleWaitForActive() {
	// Create a godo client.
	client := godo.NewFromToken("dop_v1_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")

	// Create a Droplet.
	droplet, resp, err := client.Droplets.Create(context.Background(), &godo.DropletCreateRequest{
		Name:   "test-droplet",
		Region: "nyc3",
		Size:   "s-1vcpu-1gb",
		Image: godo.DropletCreateImage{
			Slug: "ubuntu-20-04-x64",
		},
	})
	if err != nil {
		log.Fatalf("failed to create droplet: %v\n", err)
	}

	// Find the Droplet create action, then wait for it to complete.
	for _, action := range resp.Links.Actions {
		if action.Rel == "create" {
			// Block until the action is complete.
			if err := WaitForActive(context.Background(), client, action.HREF); err != nil {
				log.Fatalf("error waiting for droplet to become active: %v\n", err)
			}
		}
	}

	fmt.Println(droplet.Name)
}
