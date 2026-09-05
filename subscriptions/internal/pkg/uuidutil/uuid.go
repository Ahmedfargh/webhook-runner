package uuidutil

import (
	"github.com/google/uuid"
)

// ParseOrHash converts any string into a deterministic valid UUID.
// If the string is already a valid UUID, it returns it directly.
// Otherwise, it creates a deterministic UUID using SHA1 (v5) hashing.
func ParseOrHash(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	if id, err := uuid.Parse(s); err == nil {
		return id
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(s))
}
