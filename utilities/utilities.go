package utilities

import (
	uuid "github.com/satori/go.uuid"
)

// NewGUID creates a new guid for record keys
func NewGUID() string {
	return uuid.Must(uuid.NewV4()).String()
}
