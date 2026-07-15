package utils

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
)

func GenerateSecretKey(password, username string) string {
	salt := []byte(username)

	rawKey, err := pbkdf2.Key(sha256.New, password, salt, 4096, 32)
	if err != nil {
		slog.Error("failed to derive pbkdf2 key", "err", err)
		return ""
	}

	return hex.EncodeToString(rawKey)
}
