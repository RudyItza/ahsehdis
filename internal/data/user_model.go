package data

import (
	"database/sql"
	"errors"
)

// Custom errors for specific user-related database scenarios
var (
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrRecordNotFound = errors.New("record not found")
)

// UserModel wraps a sql.DB connection pool for working with user data
type UserModel struct {
	DB *sql.DB
}

// Insert adds a new user to the database and sets the ID, created_at, and updated_at fields
func (m *UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`
	// Execute the insert query and scan returned fields into the user struct
	err := m.DB.QueryRow(query, user.Email, user.PasswordHash).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	// Check for duplicate email error based on PostgreSQL constraint violation
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

// GetByEmail fetches a user from the database by email address
func (m *UserModel) GetByEmail(email string) (*User, error) {
	// Query the database and scan the row into a User struct
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1`

	var user User
	err := m.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	// If no rows returned, wrap and return ErrRecordNotFound
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByID fetches a user from the database by their ID
func (m *UserModel) GetByID(id int) (*User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1`
	// Execute the query and populate the user struct
	var user User
	err := m.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	// Handle case where user ID does not exist in the database
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}
