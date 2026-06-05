package provider

import (
	"context"

	"syntheticvision/internal/config"
)

// GenerateRequest describes a single image to synthesize.
type GenerateRequest struct {
	Mode           string
	Prompt         string
	NegativePrompt string
	Style          string
	Width          int
	Height         int
	Seed           int64
	SourceImage    *InputImage
	MaskImage      *InputImage
}

// InputImage carries an uploaded image used as a reference or edit mask.
type InputImage struct {
	Data     []byte
	MimeType string
	Filename string
}

// GenerateResult carries the produced image bytes and its MIME type.
type GenerateResult struct {
	Image    []byte
	MimeType string
}

// Provider abstracts an image-generation backend.
type Provider interface {
	Name() string
	Generate(ctx context.Context, req GenerateRequest) (*GenerateResult, error)
}

// New returns the provider selected by cfg.ImageProvider. It falls back to the
// deterministic mock provider for any value other than "openai".
func New(cfg config.Config) Provider {
	switch cfg.ImageProvider {
	case "openai":
		return NewOpenAIProvider(cfg)
	default:
		return NewMockProvider(cfg)
	}
}
