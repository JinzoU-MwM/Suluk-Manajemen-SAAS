package ai

import (
	"context"

	"github.com/jamaah-in/v2/internal/shared/config"
)

// Generator is the provider-agnostic text-generation interface used by the
// accounting copilot. Both the Gemini *Client and the OpenCode generator satisfy it.
type Generator interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
	Available() bool
}

// New returns the configured text Generator, or nil when the selected provider's
// API key is empty (callers treat nil as "AI unavailable", a graceful non-fatal
// state). The key is checked BEFORE constructing, so a nil concrete pointer is
// never wrapped into a non-nil interface.
func New(cfg *config.Config) Generator {
	switch cfg.AI.Provider {
	case "opencode":
		if cfg.AI.OpenCodeAPIKey == "" {
			return nil
		}
		return newOpenCode(cfg.AI.OpenCodeAPIKey, cfg.AI.OpenCodeModel, cfg.AI.OpenCodeBaseURL)
	default: // "gemini"
		if cfg.Gemini.APIKey == "" {
			return nil
		}
		return newGemini(cfg.Gemini.APIKey)
	}
}
