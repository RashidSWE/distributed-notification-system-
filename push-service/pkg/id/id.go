package id

import (
	"github.com/google/uuid"
)

// generates a new UUID string
func Generate() string {
	return uuid.New().String()
}
