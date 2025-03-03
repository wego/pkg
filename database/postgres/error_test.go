package postgres_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/database/postgres"
)

func TestIsLockError(t *testing.T) {
	// Test case: Error is not a *pgconn.PgError
	err := errors.New("some error")
	assert.False(t, postgres.IsLockError(err))

	// Test case: Error is a *pgconn.PgError with a different code
	pgErr := &pgconn.PgError{Code: "12345"}
	assert.False(t, postgres.IsLockError(pgErr))

	// Test case: Error is a *pgconn.PgError with the lock error code
	pgErr = &pgconn.PgError{Code: "55P03"}
	assert.True(t, postgres.IsLockError(pgErr))
}

func TestIsUniqueConstraintError(t *testing.T) {
	// Test case: Error is not a *pgconn.PgError
	err := errors.New("some error")
	assert.False(t, postgres.IsUniqueConstraintError(err))

	// Test case: Error is a *pgconn.PgError with a different code
	pgErr := &pgconn.PgError{Code: "12345"}
	assert.False(t, postgres.IsUniqueConstraintError(pgErr))

	// Test case: Error is a *pgconn.PgError with the unique constraint error code
	pgErr = &pgconn.PgError{Code: "23505"}
	assert.True(t, postgres.IsUniqueConstraintError(pgErr))
}
