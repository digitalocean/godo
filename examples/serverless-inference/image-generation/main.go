// Command image-generation generates a single image via the Serverless Inference API
// and writes it to image.png.
package main

import (
	"context"
	"encoding/base64"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))

	ctx := context.Background()

	prompt := "A cute robot in a forest of trees."

	print("> ")
	println(prompt)
	println()

	image, _, err := client.Images.Generate(ctx, &godo.ImageGenerateParams{
		Model:  "stable-diffusion-3.5-large",
		Prompt: prompt,
		N:      1,
	})
	if err != nil {
		panic(err)
	}

	println("Image Base64 Length:")
	println(len(image.Data[0].B64JSON))
	println()

	imageBytes, err := base64.StdEncoding.DecodeString(image.Data[0].B64JSON)
	if err != nil {
		panic(err)
	}

	dest := "./image.png"
	println("Writing image to " + dest)
	if err := os.WriteFile(dest, imageBytes, 0644); err != nil {
		panic(err)
	}
}
