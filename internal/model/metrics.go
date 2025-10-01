// Package models содержит внутренние типы и структуры для работы с метриками.
// Здесь определены типы Gauge и Counter для хранения значений,
// а структура Metrics используется для API и сериализации.
package models

type Gauge float64
type Counter int64

type MetricType string

const (
	TypeCounter MetricType = "counter"
	TypeGauge   MetricType = "gauge"
)

var ValidMetricTypes = map[MetricType]bool{
	TypeCounter: true,
	TypeGauge:   true,
}

// Metrics NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType MetricType   `json:"type"`
	Delta *Counter `json:"delta,omitempty"` // для counter
	Value *Gauge   `json:"value,omitempty"` // для gauge
	Hash  string   `json:"hash,omitempty"`
}

func IsValidMetricType(metType MetricType) bool {
	return ValidMetricTypes[metType]
}
