package repository

import (
	"context"
	"errors"
	"testing"
)

// ── withRetry ─────────────────────────────────────────────────────────────────

func TestWithRetry_Success(t *testing.T) {
	calls := 0
	err := withRetry(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_NonRetriableError(t *testing.T) {
	// Не-retriable ошибка — немедленный возврат без повторов.
	err := withRetry(context.Background(), func() error {
		return errors.New("non-retriable error")
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestWithRetry_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // отменяем сразу

	err := withRetry(ctx, func() error {
		return errors.New("any error")
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestWithRetry_SuccessAfterFirstError(t *testing.T) {
	attempts := 0
	err := withRetry(context.Background(), func() error {
		attempts++
		if attempts == 1 {
			return errors.New("first attempt error")
		}
		return nil
	})
	// Первая ошибка не pg-retriable → сразу возвращает ошибку.
	if err == nil {
		t.Error("expected error for non-retriable first failure")
	}
}

// ── NewDBMetricStorage с плохим DSN ──────────────────────────────────────────

func TestNewDBMetricStorage_InvalidDSN(t *testing.T) {
	_, err := NewDBMetricStorage("invalid://bad-dsn")
	if err == nil {
		t.Error("expected error for invalid DSN")
	}
}
