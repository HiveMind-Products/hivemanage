package crypt

import (
	"strings"
	"testing"
)

func TestGeneratorLengths(t *testing.T) {
	cases := []struct {
		name string
		gen  func() (string, error)
		want int
	}{
		{"GenerateSessionID", GenerateSessionID, 24},
		{"GenerateApiKey", GenerateApiKey, 32},
		{"GenerateFilename", GenerateFilename, 24},
		{"GeneratePrimaryKey", GeneratePrimaryKey, 16},
		{"GeneratePassword", GeneratePassword, 16},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := tc.gen()
			if err != nil {
				t.Fatalf("%s returned error: %v", tc.name, err)
			}
			if len(id) != tc.want {
				t.Fatalf("%s length = %d, want %d (%q)", tc.name, len(id), tc.want, id)
			}
		})
	}
}

func TestGeneratorCharset(t *testing.T) {
	const alnumUnderscore = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"
	const alnum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const passwordSet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!@#$%"

	cases := []struct {
		name    string
		gen     func() (string, error)
		charset string
	}{
		{"GenerateSessionID", GenerateSessionID, alnumUnderscore},
		{"GenerateApiKey", GenerateApiKey, alnumUnderscore},
		{"GenerateFilename", GenerateFilename, alnum},
		{"GeneratePrimaryKey", GeneratePrimaryKey, alnum},
		{"GeneratePassword", GeneratePassword, passwordSet},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < 200; i++ {
				id, err := tc.gen()
				if err != nil {
					t.Fatalf("%s returned error: %v", tc.name, err)
				}
				for _, r := range id {
					if !strings.ContainsRune(tc.charset, r) {
						t.Fatalf("%s produced out-of-charset rune %q in %q", tc.name, r, id)
					}
				}
			}
		})
	}
}

func TestGeneratorUniqueness(t *testing.T) {
	cases := []struct {
		name string
		gen  func() (string, error)
	}{
		{"GenerateSessionID", GenerateSessionID},
		{"GenerateApiKey", GenerateApiKey},
		{"GenerateFilename", GenerateFilename},
		{"GeneratePrimaryKey", GeneratePrimaryKey},
		{"GeneratePassword", GeneratePassword},
	}

	const iterations = 1000
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seen := make(map[string]struct{}, iterations)
			for i := 0; i < iterations; i++ {
				id, err := tc.gen()
				if err != nil {
					t.Fatalf("%s returned error: %v", tc.name, err)
				}
				if _, dup := seen[id]; dup {
					t.Fatalf("%s produced duplicate value %q after %d iterations", tc.name, id, i)
				}
				seen[id] = struct{}{}
			}
		})
	}
}
