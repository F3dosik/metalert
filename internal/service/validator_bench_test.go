package service_test

import (
	"testing"

	"github.com/F3dosik/metalert.git/internal/service"
	"github.com/F3dosik/metalert.git/pkg/models"
)

func BenchmarkCheckAndParseValue_Gauge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		service.CheckAndParseValue(models.TypeGauge, "Alloc", "123.45")
	}
}

func BenchmarkCheckAndParseValue_Counter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		service.CheckAndParseValue(models.TypeCounter, "Requests", "42")
	}
}

func BenchmarkCheckAndParseValue_InvalidType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		service.CheckAndParseValue("unknown", "Alloc", "123.45")
	}
}
