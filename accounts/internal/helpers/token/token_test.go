package token_test

import (
	"encoding/hex"
	"testing"

	"accounts/internal/helpers/token"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSecureToken(t *testing.T) {
	// Test default 32 bytes (64 hex characters)
	tok, err := token.GenerateSecureToken(32)
	require.NoError(t, err)
	assert.Len(t, tok, 64)

	decoded, err := hex.DecodeString(tok)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)

	// Test uniqueness
	tok2, err := token.GenerateSecureToken(32)
	require.NoError(t, err)
	assert.NotEqual(t, tok, tok2)
}
