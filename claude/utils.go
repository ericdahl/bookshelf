package main

import (
	"crypto/rand"
	"encoding/hex"
)

// generateID creates a random ID for database records
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
