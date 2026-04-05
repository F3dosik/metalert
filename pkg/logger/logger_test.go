package logger

import "testing"

func TestNewLogger_Development(t *testing.T) {
	l, s := NewLogger(ModeDevelopment)
	if l == nil || s == nil {
		t.Fatal("NewLogger returned nil")
	}
	_ = l.Sync()
}

func TestNewLogger_Production(t *testing.T) {
	l, s := NewLogger(ModeProduction)
	if l == nil || s == nil {
		t.Fatal("NewLogger returned nil")
	}
	_ = l.Sync()
}

func TestNewLogger_UnknownMode(t *testing.T) {
	// Неизвестный режим использует development fallback.
	l, s := NewLogger("unknown")
	if l == nil || s == nil {
		t.Fatal("NewLogger returned nil for unknown mode")
	}
	_ = l.Sync()
}
