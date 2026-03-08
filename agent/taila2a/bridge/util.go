package main

import (
	"os"
	"strconv"
)

// getenv returns the value of an environment variable or a fallback
func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getenvInt returns the integer value of an environment variable or a fallback
func getenvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
