package pgerrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsRetriable_RegularError(t *testing.T) {
	if IsRetriable(errors.New("some error")) {
		t.Error("regular error should not be retriable")
	}
}

func TestIsRetriable_NilError(t *testing.T) {
	if IsRetriable(nil) {
		t.Error("nil should not be retriable")
	}
}

func TestIsRetriable_ConnectionException(t *testing.T) {
	err := &pgconn.PgError{Code: "08000"} // ConnectionException
	if !IsRetriable(err) {
		t.Error("08000 should be retriable")
	}
}

func TestIsRetriable_ConnectionDoesNotExist(t *testing.T) {
	err := &pgconn.PgError{Code: "08003"} // ConnectionDoesNotExist
	if !IsRetriable(err) {
		t.Error("08003 should be retriable")
	}
}

func TestIsRetriable_ConnectionFailure(t *testing.T) {
	err := &pgconn.PgError{Code: "08006"} // ConnectionFailure
	if !IsRetriable(err) {
		t.Error("08006 should be retriable")
	}
}

func TestIsRetriable_CannotConnectNow(t *testing.T) {
	err := &pgconn.PgError{Code: "57P03"} // CannotConnectNow
	if !IsRetriable(err) {
		t.Error("57P03 should be retriable")
	}
}

func TestIsRetriable_NonRetriablePgError(t *testing.T) {
	err := &pgconn.PgError{Code: "23505"} // UniqueViolation — не retriable
	if IsRetriable(err) {
		t.Error("unique violation should not be retriable")
	}
}

func TestIsRetriable_WrappedPgError(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "08000"}
	wrapped := fmt.Errorf("wrapped: %w", pgErr)
	if !IsRetriable(wrapped) {
		t.Error("wrapped connection error should be retriable")
	}
}
