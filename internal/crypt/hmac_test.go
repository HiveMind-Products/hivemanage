package crypt

import (
	"encoding/hex"
	"testing"
)

func TestComputeHMACDeterministic(t *testing.T) {
	cases := []struct {
		name    string
		secret  string
		message string
	}{
		{"basic", "secret", "message"},
		{"empty-message", "secret", ""},
		{"empty-secret", "", "message"},
		{"token-like", "hmac-secret-key", "abc123APITOKEN"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := ComputeHMAC(tc.secret, tc.message)
			b := ComputeHMAC(tc.secret, tc.message)
			if a != b {
				t.Fatalf("ComputeHMAC not deterministic: %q != %q", a, b)
			}
		})
	}
}

func TestComputeHMACDiffersBySecret(t *testing.T) {
	msg := "same-message"
	a := ComputeHMAC("secret-a", msg)
	b := ComputeHMAC("secret-b", msg)
	if a == b {
		t.Fatalf("expected different HMAC for different secrets")
	}
}

func TestComputeHMACDiffersByMessage(t *testing.T) {
	secret := "same-secret"
	a := ComputeHMAC(secret, "message-a")
	b := ComputeHMAC(secret, "message-b")
	if a == b {
		t.Fatalf("expected different HMAC for different messages")
	}
}

func TestComputeHMACOutputIsConstantLengthHex(t *testing.T) {
	// SHA-256 HMAC => 32 bytes => 64 hex characters, regardless of input size.
	cases := []struct {
		secret  string
		message string
	}{
		{"", ""},
		{"s", "m"},
		{"longer-secret-value", "a-much-longer-message-payload-than-the-secret"},
	}

	for _, tc := range cases {
		out := ComputeHMAC(tc.secret, tc.message)
		if len(out) != 64 {
			t.Fatalf("expected 64 hex chars, got %d for secret=%q message=%q", len(out), tc.secret, tc.message)
		}
		if _, err := hex.DecodeString(out); err != nil {
			t.Fatalf("output is not valid hex: %v", err)
		}
	}
}
