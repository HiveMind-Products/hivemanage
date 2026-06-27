package crypt

import "testing"

func TestHashPasswordRoundTrip(t *testing.T) {
	cases := []struct {
		name     string
		password string
	}{
		{"simple", "hunter2"},
		{"empty", ""},
		{"unicode", "pâsswörd-🔐"},
		{"long", "this-is-a-fairly-long-password-with-symbols-!@#$%^&*()"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := HashPassword(tc.password)
			if err != nil {
				t.Fatalf("HashPassword returned error: %v", err)
			}
			if hash == tc.password {
				t.Fatalf("hash must not equal plaintext password")
			}
			if err := ComparePassword(hash, tc.password); err != nil {
				t.Fatalf("ComparePassword rejected correct password: %v", err)
			}
		})
	}
}

func TestComparePasswordRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	cases := []string{"wrong-horse", "correct-horse ", "Correct-Horse", "", "correct-hors"}
	for _, wrong := range cases {
		if err := ComparePassword(hash, wrong); err == nil {
			t.Fatalf("ComparePassword accepted wrong password %q", wrong)
		}
	}
}

func TestHashPasswordUsesSalt(t *testing.T) {
	// Two hashes of the same password must differ because of a random salt,
	// yet both must verify against the original password.
	h1, err := HashPassword("same-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	h2, err := HashPassword("same-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if h1 == h2 {
		t.Fatalf("expected distinct hashes for identical passwords (missing salt)")
	}
	if err := ComparePassword(h1, "same-password"); err != nil {
		t.Fatalf("first hash failed to verify: %v", err)
	}
	if err := ComparePassword(h2, "same-password"); err != nil {
		t.Fatalf("second hash failed to verify: %v", err)
	}
}
