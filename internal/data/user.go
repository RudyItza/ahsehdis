package data

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system, including authentication details and timestamps
type User struct {
	ID           int
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SetPassword hashes a plaintext password using bcrypt and stores the result in the PasswordHash field
func (u *User) SetPassword(plaintext string) error {
	// bcrypt.GenerateFromPassword hashes the plaintext with a cost of 12 (relatively secure)
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err // Return the error if hashing fails
	}
	u.PasswordHash = string(hash) // Store the resulting hash in the User struct
	return nil
}

// MatchesPassword compares the stored hash with a plaintext password to verify authentication
func (u *User) MatchesPassword(plaintext string) error {
	// bcrypt.CompareHashAndPassword returns nil if the password matches the hash
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plaintext))
}
