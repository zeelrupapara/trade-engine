package utils

import (
	"github.com/google/uuid"
)

func GenerateUUID() string {
	// using google uuid
	return uuid.New().String()
}
