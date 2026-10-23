package utils

import (
	"log"
	"os"
	"strconv"
)

// checkerCountFromEnvironment returns the checker count requested through
// OXLINT_TSGOLINT_CHECKERS, or nil when unset or invalid. It takes precedence
// over `compilerOptions.checkers`, so memory-constrained runs can use fewer
// checkers without editing a tsconfig that other tools also read.
func checkerCountFromEnvironment() *int {
	value := os.Getenv("OXLINT_TSGOLINT_CHECKERS")
	if value == "" {
		return nil
	}
	count, err := strconv.Atoi(value)
	if err != nil || count < 1 {
		log.Printf("Ignoring OXLINT_TSGOLINT_CHECKERS=%q: expected a positive integer", value)
		return nil
	}
	return &count
}
