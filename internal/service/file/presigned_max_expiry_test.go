package file

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

// decodeExp pulls the unix expiry out of a presigned token's claims segment.
func decodeExp(t *testing.T, token string) int64 {
	t.Helper()
	encoded, _, ok := strings.Cut(token, ".")
	if !ok {
		t.Fatalf("token missing signature separator: %q", token)
	}
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("failed to decode claims: %v", err)
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("failed to unmarshal claims: %v", err)
	}
	return claims.Exp
}

func TestCreatePresignedTokenClampsExpiry(t *testing.T) {
	t.Setenv("API_TOKEN_HMAC_SECRET", "fixed-test-secret")
	svc := &Service{}

	cases := []struct {
		name      string
		requested time.Duration // offset from now for expiresAt
	}{
		{"far-future-clamped-to-max", 365 * 24 * time.Hour},
		{"just-over-max-clamped", MaxPresignedExpiry + time.Hour},
		{"within-max-preserved", time.Hour},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := time.Now()
			token := svc.CreatePresignedToken("org123", before.Add(tc.requested), "avatars")
			after := time.Now()

			exp := decodeExp(t, token)
			maxExp := after.Add(MaxPresignedExpiry).Unix()
			if exp > maxExp {
				t.Fatalf("exp %d exceeds clamp ceiling now+MaxPresignedExpiry %d", exp, maxExp)
			}
			// The token must still be valid (not already expired) right now.
			if _, _, err := svc.ResolvePresignedToken(token); err != nil {
				t.Fatalf("freshly minted token should be valid: %v", err)
			}
		})
	}
}

func TestCreatePresignedTokenPastExpiryGetsDefaultWindow(t *testing.T) {
	t.Setenv("API_TOKEN_HMAC_SECRET", "fixed-test-secret")
	svc := &Service{}

	// A request for an already-past expiry must not produce an expired token; it
	// is bumped to the default window instead.
	token := svc.CreatePresignedToken("org123", time.Now().Add(-time.Hour), "")
	if _, _, err := svc.ResolvePresignedToken(token); err != nil {
		t.Fatalf("past-expiry request should yield a valid default-window token: %v", err)
	}
	exp := decodeExp(t, token)
	if exp <= time.Now().Unix() {
		t.Fatalf("expected future expiry, got %d", exp)
	}
}

func TestVerifyPresignedTokenRejectsNonPositiveExp(t *testing.T) {
	const secret = "fixed-test-secret"
	for _, exp := range []int64{0, -1, -1000} {
		token := BuildPresignedToken(secret, "org123", "avatars", exp)
		if _, _, err := VerifyPresignedToken(secret, token, time.Now().Unix()); !errors.Is(err, ErrPresignedInvalid) {
			t.Fatalf("exp=%d: expected ErrPresignedInvalid, got %v", exp, err)
		}
	}
}
