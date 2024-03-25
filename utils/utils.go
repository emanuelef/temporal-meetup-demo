package utils

import "os"

func GetEnv(envVar, defaultValue string) string {
	val, exists := os.LookupEnv(envVar)
	if !exists {
		return defaultValue
	}
	return val
}
