package util

import (
	"context"
	"fmt"
	"log"

	"github.com/digitalocean/godo"
)

func ExampleWaitForAvailable() {
	// Create a godo client.
	client := godo.NewFromToken("dop_v1_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")

	// Create an Image.
	image, resp, err := client.Images.Create(context.Background(), &godo.CustomImageCreateRequest{
		Name:   "test-image",
		Url:    "https://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.vmdk",
		Region: "nyc3",
	})
	if err != nil {
		log.Fatalf("failed to create image: %v\n", err)
	}

	// Find the Image create action, then wait for it to complete.
	for _, action := range resp.Links.Actions {
		if action.Rel == "create" {
			// Block until the action is complete.
			if err := WaitForAvailable(context.Background(), client, action.HREF); err != nil {
				log.Fatalf("error waiting for image to become active: %v\n", err)
			}
		}
	}

	fmt.Println(image.Name)
}
