package postgres

import "github.com/jackc/pgconn"

// IsLockError checks if the error is a lock error
func IsLockError(err error) bool {
	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return false
	}

	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	// 55P03 -> lock_not_available
	return pgErr.Code == "55P03"
}

// IsUniqueConstraintError checks if the error is a unique constraint error
func IsUniqueConstraintError(err error) bool {
	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return false
	}

	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	// 23505 -> unique_violation
	return pgErr.Code == "23505"
}
