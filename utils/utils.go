package utils

import (
	"os"

	"github.com/google/uuid"
)

func GetEnv(envVar, defaultValue string) string {
	val, exists := os.LookupEnv(envVar)
	if !exists {
		return defaultValue
	}
	return val
}

func GenerateUUID() string {
	id := uuid.New()
	return id.String()
}