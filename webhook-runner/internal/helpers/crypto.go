package helpers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateRandomHex generates secure random hex characters of specified byte length
func GenerateRandomHex(bytesLen int) (string, error) {
	b := make([]byte, bytesLen)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateAppID generates a public App ID (e.g. app_live_3f4e8b...)
func GenerateAppID() string {
	randomPart, _ := GenerateRandomHex(12)
	return fmt.Sprintf("app_live_%s", randomPart)
}

// GenerateAppSecret generates a private App Secret (e.g. sec_...)
func GenerateAppSecret() string {
	randomPart, _ := GenerateRandomHex(24)
	return fmt.Sprintf("sec_%s", randomPart)
}

// GenerateWebhookSecret generates a Webhook HMAC signing secret (e.g. whsec_...)
func GenerateWebhookSecret() string {
	randomPart, _ := GenerateRandomHex(20)
	return fmt.Sprintf("whsec_%s", randomPart)
}

// ComputeHMACSHA256 generates a cryptographic HMAC-SHA256 signature for payload
func ComputeHMACSHA256(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
