package http

import (
	"testing"
)

func TestAllowedOriginUsesAllowlist(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com")

	allowed, err := allowedOrigin("https://admin.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatalf("expected configured origin to be allowed")
	}

	allowed, err = allowedOrigin("https://evil.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatalf("expected unknown origin to be rejected")
	}
}
