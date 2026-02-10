package utils

import (
	"crypto/rand"
	"encoding/json"
)

const slugChars = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandomSlug returns a URL-safe random slug of length n (e.g. for group share links).
func RandomSlug(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = slugChars[int(b[i])%len(slugChars)]
	}
	return string(b)
}

func Dump(data interface{}) string {
	btl, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(btl)
}
