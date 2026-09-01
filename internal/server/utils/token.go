package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken returns a deterministic SHA-256 hex digest of token.
// Use for storing/looking up high-entropy refresh tokens (not reversible like base64).
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
