// Package models содержит внутренние типы и структуры для работы с метриками.
// Здесь определены типы Gauge и Counter для хранения значений,
// а структура Metrics используется для API и сериализации.
package models

type Gauge float64
type Counter int64

const (
	CounterType = "counter"
	GaugeType   = "gauge"
)

// Metrics NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
