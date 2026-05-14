// Command image-streaming streams partial-image events via the Serverless Inference API.
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))
	ctx := context.Background()

	model := os.Getenv("DIGITALOCEAN_INFERENCE_IMAGE_MODEL")
	if model == "" {
		model = "openai-gpt-image-1"
	}

	fmt.Println("Starting image streaming example...")

	stream, _, err := client.ImageGenerations.GenerateStreaming(ctx, &godo.ImageGenerateParams{
		Model:         model,
		Prompt:        "A cute baby sea otter",
		N:             1,
		Size:          godo.PtrTo("1024x1024"),
		PartialImages: godo.PtrTo(3),
	})
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	for stream.Next() {
		event := stream.Current()

		switch event.Type {
		case "image_generation.partial_image":
			fmt.Printf("  Partial image %d/3 received\n", event.PartialImageIndex+1)
			fmt.Printf("   Size: %d characters (base64)\n", len(event.B64JSON))

			filename := fmt.Sprintf("partial_%d.png", event.PartialImageIndex+1)
			if err := saveBase64Image(event.B64JSON, filename); err != nil {
				panic(fmt.Errorf("failed to save partial image: %w", err))
			}
			absPath, _ := filepath.Abs(filename)
			fmt.Printf("   Saved to: %s\n", absPath)
		case "image_generation.completed":
			fmt.Printf("\nFinal image completed!\n")
			fmt.Printf("   Size: %d characters (base64)\n", len(event.B64JSON))

			filename := "final_image.png"
			if err := saveBase64Image(event.B64JSON, filename); err != nil {
				panic(fmt.Errorf("failed to save final image: %w", err))
			}
			absPath, _ := filepath.Abs(filename)
			fmt.Printf("   Saved to: %s\n", absPath)

		default:
			fmt.Printf("Received unknown event type: %+v\n", event)
		}
	}

	if err := stream.Err(); err != nil {
		panic(fmt.Errorf("error during streaming: %w", err))
	}
}

func saveBase64Image(b64Data, filename string) error {
	imageData, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	if err := os.WriteFile(filename, imageData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
