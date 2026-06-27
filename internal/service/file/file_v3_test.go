package file

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestDecodeBase64Payload(t *testing.T) {
	want := []byte("hello world")
	std := base64.StdEncoding.EncodeToString(want)

	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"plain std base64", std, false},
		{"data uri prefix", "data:text/plain;base64," + std, false},
		{"url-safe base64", base64.RawURLEncoding.EncodeToString(want), false},
		{"whitespace padded", "  " + std + "  ", false},
		{"empty", "", true},
		{"data uri without base64 marker", "data:text/plain,hello", true},
		{"garbage", "!!!not base64!!!", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecodeBase64Payload(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none (decoded %q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != string(want) {
				t.Fatalf("decoded %q, want %q", got, want)
			}
		})
	}
}

func TestParseMetadataJSON(t *testing.T) {
	t.Run("empty yields nil", func(t *testing.T) {
		m, err := ParseMetadataJSON("")
		if err != nil || m != nil {
			t.Fatalf("expected (nil,nil), got (%v,%v)", m, err)
		}
	})
	t.Run("valid object", func(t *testing.T) {
		m, err := ParseMetadataJSON(`{"a":"b","n":1}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m["a"] != "b" {
			t.Fatalf("expected a=b, got %v", m["a"])
		}
	})
	t.Run("invalid json errors", func(t *testing.T) {
		if _, err := ParseMetadataJSON(`{not json`); err == nil {
			t.Fatal("expected error for invalid json")
		}
	})
}

func TestGenerateFileKeyV3(t *testing.T) {
	org := "org123"

	t.Run("path and filename", func(t *testing.T) {
		key, err := generateFileKeyV3(org, "avatars", "pic.png", ".png")
		if err != nil {
			t.Fatal(err)
		}
		if key != "org123/avatars/pic.png" {
			t.Fatalf("got %q", key)
		}
	})

	t.Run("random name when empty", func(t *testing.T) {
		key, err := generateFileKeyV3(org, "", "", ".jpg")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(key, "org123/") || !strings.HasSuffix(key, ".jpg") {
			t.Fatalf("got %q", key)
		}
	})

	t.Run("appends extension when filename has none", func(t *testing.T) {
		key, err := generateFileKeyV3(org, "", "report", ".pdf")
		if err != nil {
			t.Fatal(err)
		}
		if key != "org123/report.pdf" {
			t.Fatalf("got %q", key)
		}
	})

	t.Run("strips path traversal in folder", func(t *testing.T) {
		key, err := generateFileKeyV3(org, "../../etc", "x.txt", ".txt")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(key, "org123/") || strings.Contains(key, "..") {
			t.Fatalf("traversal not stripped: %q", key)
		}
	})

	t.Run("strips directory from filename", func(t *testing.T) {
		key, err := generateFileKeyV3(org, "", "../../secret.txt", ".txt")
		if err != nil {
			t.Fatal(err)
		}
		if key != "org123/secret.txt" {
			t.Fatalf("got %q", key)
		}
	})
}

func TestComposeFileURLs(t *testing.T) {
	key := "org/abc.png"
	cases := []struct {
		name         string
		bucketDomain string
		defaultURL   string
		wantURL      string
		wantOriginal string
	}{
		{"cdn + default", "https://cdn.example.com", "https://s3.example.com/bucket/org/abc.png", "https://cdn.example.com/org/abc.png", "https://s3.example.com/bucket/org/abc.png"},
		{"no cdn falls back to default", "", "https://s3.example.com/bucket/org/abc.png", "https://s3.example.com/bucket/org/abc.png", "https://s3.example.com/bucket/org/abc.png"},
		{"no default falls back to cdn", "https://cdn.example.com", "", "https://cdn.example.com/org/abc.png", "https://cdn.example.com/org/abc.png"},
		{"trailing slash trimmed", "https://cdn.example.com/", "https://s3/x", "https://cdn.example.com/org/abc.png", "https://s3/x"},
		{"both empty", "", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url, original := composeFileURLs(tc.bucketDomain, tc.defaultURL, key)
			if url != tc.wantURL || original != tc.wantOriginal {
				t.Fatalf("got (%q,%q), want (%q,%q)", url, original, tc.wantURL, tc.wantOriginal)
			}
		})
	}
}

func TestNormalizeListParams(t *testing.T) {
	cases := []struct {
		page, limit             int
		wantPage, wantLim, wOff int
	}{
		{0, 0, 1, DefaultListLimit, 0},
		{1, 50, 1, 50, 0},
		{3, 20, 3, 20, 40},
		{2, 1000, 2, MaxListLimit, MaxListLimit},
		{-5, -5, 1, DefaultListLimit, 0},
	}
	for _, tc := range cases {
		p, l, off := NormalizeListParams(tc.page, tc.limit)
		if p != tc.wantPage || l != tc.wantLim || off != tc.wOff {
			t.Fatalf("NormalizeListParams(%d,%d)=(%d,%d,%d), want (%d,%d,%d)", tc.page, tc.limit, p, l, off, tc.wantPage, tc.wantLim, tc.wOff)
		}
	}
}
