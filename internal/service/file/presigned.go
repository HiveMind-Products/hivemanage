package file

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/fivemanage/lite/internal/crypt"
)

// DefaultPresignedExpiry is the validity window for a presigned upload URL when
// the caller does not supply an explicit expiresAt.
const DefaultPresignedExpiry = 15 * time.Minute

var (
	// ErrPresignedInvalid is returned for a malformed or tampered presigned token.
	ErrPresignedInvalid = errors.New("invalid presigned token")
	// ErrPresignedExpired is returned for an expired presigned token.
	ErrPresignedExpired = errors.New("presigned token has expired")
)

type presignedClaims struct {
	Org  string `json:"org"`
	Exp  int64  `json:"exp"`
	Path string `json:"path,omitempty"`
}

// BuildPresignedToken signs a presigned-upload token encoding the organization,
// expiry (unix seconds) and optional path. Format: base64url(claims).hex(hmac).
func BuildPresignedToken(secret, org, path string, exp int64) string {
	payload, _ := json.Marshal(presignedClaims{Org: org, Exp: exp, Path: path})
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return encoded + "." + crypt.ComputeHMAC(secret, encoded)
}

// VerifyPresignedToken validates a token's signature and expiry, returning the
// encoded organization and path. now is unix seconds.
func VerifyPresignedToken(secret, token string, now int64) (org string, path string, err error) {
	encoded, sig, ok := strings.Cut(token, ".")
	if !ok || encoded == "" || sig == "" {
		return "", "", ErrPresignedInvalid
	}
	expected := crypt.ComputeHMAC(secret, encoded)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(sig)) != 1 {
		return "", "", ErrPresignedInvalid
	}
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", ErrPresignedInvalid
	}
	var claims presignedClaims
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Org == "" {
		return "", "", ErrPresignedInvalid
	}
	if claims.Exp != 0 && now > claims.Exp {
		return "", "", ErrPresignedExpired
	}
	return claims.Org, claims.Path, nil
}

// CreatePresignedToken issues a presigned-upload token for the organization.
func (s *Service) CreatePresignedToken(org string, expiresAt time.Time, path string) string {
	return BuildPresignedToken(presignedSecret(), org, path, expiresAt.Unix())
}

// ResolvePresignedToken validates a presigned-upload token and returns the
// organization and embedded path.
func (s *Service) ResolvePresignedToken(token string) (org string, path string, err error) {
	return VerifyPresignedToken(presignedSecret(), token, time.Now().Unix())
}

func presignedSecret() string {
	return os.Getenv("API_TOKEN_HMAC_SECRET")
}
