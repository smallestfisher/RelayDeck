package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const passwordHashVersion = "sha256"

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return passwordHashVersion + "$" + hex.EncodeToString(salt) + "$" + hashPasswordWithSalt(password, salt), nil
}

func VerifyPassword(hash string, password string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 3 || parts[0] != passwordHashVersion {
		return false
	}
	salt, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	return hashPasswordWithSalt(password, salt) == parts[2]
}

func hashPasswordWithSalt(password string, salt []byte) string {
	buffer := make([]byte, 0, len(salt)+len(password))
	buffer = append(buffer, salt...)
	buffer = append(buffer, password...)
	sum := sha256.Sum256(buffer)
	return hex.EncodeToString(sum[:])
}
