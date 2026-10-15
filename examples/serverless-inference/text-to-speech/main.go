// Command text-to-speech calls POST /v1/audio/speech and writes the audio file.
package main

import (
	"context"
	"os"

	"github.com/digitalocean/godo"
)

func main() {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		token = os.Getenv("MODEL_ACCESS_KEY")
	}
	client := godo.NewFromToken(token)
	ctx := context.Background()

	format := os.Getenv("SPEECH_RESPONSE_FORMAT")
	if format == "" {
		format = "mp3"
	}

	speech, _, err := client.Audio.Speech.New(ctx, &godo.SpeechNewParams{
		Model:          "qwen3-tts-voicedesign",
		Input:          "Hello, world!",
		Voice:          "alloy",
		ResponseFormat: format,
		Instructions:   "Speak naturally.",
	})
	if err != nil {
		panic(err)
	}

	audioBytes, err := speech.Bytes()
	if err != nil {
		panic(err)
	}

	// Prefer the format requested; fall back if the API returned WAV (RIFF).
	if format == "mp3" && len(audioBytes) >= 4 && string(audioBytes[:4]) == "RIFF" {
		format = "wav"
	}

	dest := "./output." + format
	println("Writing audio to " + dest)
	if err := os.WriteFile(dest, audioBytes, 0644); err != nil {
		panic(err)
	}
}
