package usecase

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

// NewLobbyCode returns a short shareable code.
func NewLobbyCode() string {
	// 5 bytes -> 8 chars base32 without padding.
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	code := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	code = strings.ToUpper(code)
	// Avoid confusing characters in many fonts.
	code = strings.NewReplacer("O", "8", "I", "9", "L", "7").Replace(code)
	return code
}

func NewPlayerID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	id := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	return strings.ToLower(id)
}
