// Package password provides bcrypt password hashing and verification.
package password

import "golang.org/x/crypto/bcrypt"

// Hash returns a bcrypt hash of the plaintext password.
func Hash(plaintext string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify reports whether plaintext matches the stored bcrypt hash.
func Verify(hash, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}
