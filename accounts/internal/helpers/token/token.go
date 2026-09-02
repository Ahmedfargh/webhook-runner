package token

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateSecureToken generates a cryptographically secure random hexadecimal token of the specified byte length
func GenerateSecureToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		byteLength = 32 // 32 bytes = 64 hex characters
	}

	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}
