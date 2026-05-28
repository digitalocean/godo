// Command create-model-eval-dataset-upload-presigned-urls creates presigned
// URLs that can be used to upload model evaluation dataset files via the
// GradientAI API.
//
// Required env vars:
//   - DIGITALOCEAN_TOKEN: a DigitalOcean API token.
//   - DIGITALOCEAN_DATASET_FILE_NAME: local file name to upload.
//   - DIGITALOCEAN_DATASET_FILE_SIZE: size of the file in bytes.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	client := godo.NewFromToken(os.Getenv("DIGITALOCEAN_TOKEN"))

	required := func(name string) string {
		v := os.Getenv(name)
		if v == "" {
			fmt.Fprintf(os.Stderr, "%s is required\n", name)
			os.Exit(1)
		}
		return v
	}

	createRequest := &godo.CreateModelEvalDatasetUploadPresignedURLsRequest{
		Files: []*godo.PresignedUrlFile{
			{
				FileName: required("DIGITALOCEAN_DATASET_FILE_NAME"),
				FileSize: required("DIGITALOCEAN_DATASET_FILE_SIZE"),
			},
		},
	}

	ctx := context.Background()

	out, _, err := client.GradientAI.CreateModelEvalDatasetUploadPresignedURLs(ctx, createRequest)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Request ID: %s\n", out.RequestID)
	for _, upload := range out.Uploads {
		fmt.Printf("- file=%s key=%s url=%s\n", upload.OriginalFileName, upload.ObjectKey, upload.PresignedURL)
		if upload.ExpiresAt != nil {
			fmt.Printf("  expires at: %s\n", upload.ExpiresAt)
		}
	}
}
