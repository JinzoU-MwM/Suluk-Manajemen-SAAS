package config

import "testing"

func TestLoadAIDefaults(t *testing.T) {
	t.Setenv("AI_PROVIDER", "")
	t.Setenv("OPENCODE_API_KEY", "")
	t.Setenv("OPENCODE_MODEL", "")
	t.Setenv("OPENCODE_BASE_URL", "")
	c := Load()
	if c.AI.Provider != "opencode" {
		t.Errorf("Provider = %q, want opencode", c.AI.Provider)
	}
	if c.AI.OpenCodeModel != "claude-haiku-4-5" {
		t.Errorf("OpenCodeModel = %q, want claude-haiku-4-5", c.AI.OpenCodeModel)
	}
	if c.AI.OpenCodeBaseURL != "https://opencode.ai/zen/v1" {
		t.Errorf("OpenCodeBaseURL = %q", c.AI.OpenCodeBaseURL)
	}
	if c.AI.OpenCodeAPIKey != "" {
		t.Errorf("OpenCodeAPIKey = %q, want empty", c.AI.OpenCodeAPIKey)
	}
}

func TestLoadAIOverride(t *testing.T) {
	t.Setenv("AI_PROVIDER", "Gemini") // mixed case -> normalized
	t.Setenv("OPENCODE_API_KEY", "sk-test")
	t.Setenv("OPENCODE_MODEL", "gpt-5")
	c := Load()
	if c.AI.Provider != "gemini" {
		t.Errorf("Provider = %q, want gemini (normalized)", c.AI.Provider)
	}
	if c.AI.OpenCodeAPIKey != "sk-test" || c.AI.OpenCodeModel != "gpt-5" {
		t.Errorf("override not applied: %+v", c.AI)
	}
}
