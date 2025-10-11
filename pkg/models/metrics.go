// Package models содержит внутренние типы и структуры для работы с метриками.
// Здесь определены типы Gauge и Counter для хранения значений,
// а структура Metrics используется для API и сериализации.
package models

import "errors"

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

var (
	ErrInvalidType  = errors.New("unsupported metric type")
	ErrNoName       = errors.New("metric name not provided")
	ErrInvalidValue = errors.New("missing value for gauge metric")
	ErrInvalidDelta = errors.New("missing delta for counter metric")
)

// Metric NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metric struct {
	ID    string     `json:"id"`
	MType MetricType `json:"type"`
	Delta *Counter   `json:"delta,omitempty"` // для counter
	Value *Gauge     `json:"value,omitempty"` // для gauge
	Hash  string     `json:"hash,omitempty"`
}

func NewMetricGauge(id string, value Gauge) *Metric {
	return &Metric{
		ID:    id,
		MType: TypeGauge,
		Value: &value,
	}
}

func NewMetricCounter(id string, delta Counter) *Metric {
	return &Metric{
		ID:    id,
		MType: TypeCounter,
		Delta: &delta,
	}
}

func (met *Metric) ValidateMeta() error {
	if met.ID == "" {
		return ErrNoName
	}

	if !IsValidMetricType(met.MType) {
		return ErrInvalidType
	}

	return nil
}

func (met *Metric) ValidateValue() error {
	switch met.MType {
	case TypeGauge:
		if met.Value == nil {
			return ErrInvalidValue
		}
	case TypeCounter:
		if met.Delta == nil {
			return ErrInvalidDelta
		}
	default:
		return ErrInvalidType
	}

	return nil
}

func IsValidMetricType(metType MetricType) bool {
	return ValidMetricTypes[metType]
}
