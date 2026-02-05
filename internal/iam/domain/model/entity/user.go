package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a global user in the system
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Never expose in JSON
	FullName     *string    `json:"full_name,omitempty"`
	IsRootAdmin  bool       `json:"is_root_admin"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// User errors
var (
	ErrUserEmailRequired    = errors.New("email is required")
	ErrUserEmailInvalid     = errors.New("email format is invalid")
	ErrUserPasswordRequired = errors.New("password is required")
	ErrUserPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrInvalidCredentials   = errors.New("invalid credentials")
)

const (
	MinPasswordLength = 8
	BcryptCost        = 12
)

// NewUser creates a new user with hashed password
func NewUser(email, password string, fullName *string) (*User, error) {
	if email == "" {
		return nil, ErrUserEmailRequired
	}
	if password == "" {
		return nil, ErrUserPasswordRequired
	}
	if len(password) < MinPasswordLength {
		return nil, ErrUserPasswordTooShort
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		FullName:     fullName,
		IsRootAdmin:  false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Validate validates the user
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrUserEmailRequired
	}
	return nil
}

// VerifyPassword checks if the provided password matches
func (u *User) VerifyPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

// UpdatePassword updates the user's password
func (u *User) UpdatePassword(newPassword string) error {
	if len(newPassword) < MinPasswordLength {
		return ErrUserPasswordTooShort
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hashedPassword)
	u.UpdatedAt = time.Now()
	return nil
}

// RecordLogin records the last login time
func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// SetAsRootAdmin sets the user as a root admin
func (u *User) SetAsRootAdmin() {
	u.IsRootAdmin = true
	u.UpdatedAt = time.Now()
}

// UpdateFullName updates the user's full name
func (u *User) UpdateFullName(name string) {
	u.FullName = &name
	u.UpdatedAt = time.Now()
}
