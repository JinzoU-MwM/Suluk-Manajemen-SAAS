package service

import (
	"context"
	"testing"
)

// compile-time: the Gemini client satisfies the provider-agnostic interface.
var _ DocumentAnalyzer = (*GeminiClient)(nil)

func TestNilGeminiAnalyzeReturnsError(t *testing.T) {
	var c *GeminiClient
	if _, err := c.AnalyzeDocument(context.Background(), []byte("x"), "image/png"); err == nil {
		t.Error("nil client should return an error, not panic")
	}
	if c.Available() {
		t.Error("nil client should not be Available")
	}
}
