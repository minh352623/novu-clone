package entity

import "time"

// SystemEnvironment represents a global environment definition
type SystemEnvironment struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
