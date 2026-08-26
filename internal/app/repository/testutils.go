package repository

import (
	"testing"
)

// getTestDSN returns the test database DSN.
func getTestDSN(t *testing.T) string {
	t.Helper()
	return "postgres://postgres:postgres@localhost:5432/aegis_test?sslmode=disable"
}
