package main

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

func main() {
	fmt.Println("hit")
	client := godo.NewFromToken("my-digitalocean-api-token")
	client.Kubernetes.Get(context.Background(), "7410a3e1-4008-4284-b2e6-97f0c95266aa")
}
