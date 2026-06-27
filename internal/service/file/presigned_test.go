package file

import (
	"errors"
	"testing"
	"time"
)

func TestPresignedTokenRoundTrip(t *testing.T) {
	secret := "test-secret"
	exp := time.Now().Add(15 * time.Minute).Unix()

	token := BuildPresignedToken(secret, "org123", "avatars", exp)
	org, path, err := VerifyPresignedToken(secret, token, time.Now().Unix())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org != "org123" || path != "avatars" {
		t.Fatalf("got org=%q path=%q", org, path)
	}
}

func TestPresignedTokenNoPath(t *testing.T) {
	secret := "s"
	token := BuildPresignedToken(secret, "org", "", time.Now().Add(time.Minute).Unix())
	org, path, err := VerifyPresignedToken(secret, token, time.Now().Unix())
	if err != nil || org != "org" || path != "" {
		t.Fatalf("got org=%q path=%q err=%v", org, path, err)
	}
}

func TestPresignedTokenExpired(t *testing.T) {
	secret := "s"
	exp := time.Now().Add(-time.Minute).Unix()
	token := BuildPresignedToken(secret, "org", "", exp)
	_, _, err := VerifyPresignedToken(secret, token, time.Now().Unix())
	if !errors.Is(err, ErrPresignedExpired) {
		t.Fatalf("expected ErrPresignedExpired, got %v", err)
	}
}

func TestPresignedTokenTampered(t *testing.T) {
	secret := "s"
	token := BuildPresignedToken(secret, "org", "", time.Now().Add(time.Minute).Unix())
	tampered := token + "x"
	if _, _, err := VerifyPresignedToken(secret, tampered, time.Now().Unix()); !errors.Is(err, ErrPresignedInvalid) {
		t.Fatalf("expected ErrPresignedInvalid for tampered token, got %v", err)
	}
}

func TestPresignedTokenWrongSecret(t *testing.T) {
	token := BuildPresignedToken("secret-a", "org", "", time.Now().Add(time.Minute).Unix())
	if _, _, err := VerifyPresignedToken("secret-b", token, time.Now().Unix()); !errors.Is(err, ErrPresignedInvalid) {
		t.Fatalf("expected ErrPresignedInvalid for wrong secret, got %v", err)
	}
}

func TestPresignedTokenMalformed(t *testing.T) {
	for _, tok := range []string{"", "no-dot", ".", "abc.", ".abc"} {
		if _, _, err := VerifyPresignedToken("s", tok, time.Now().Unix()); !errors.Is(err, ErrPresignedInvalid) {
			t.Fatalf("expected ErrPresignedInvalid for %q, got %v", tok, err)
		}
	}
}
