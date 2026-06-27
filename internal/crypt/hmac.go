package crypt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// used for generating a hash of the api token
func ComputeHMAC(secret, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
