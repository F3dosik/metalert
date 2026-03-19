// Package models содержит внутренние типы и структуры для работы с метриками.
// Здесь определены типы Gauge и Counter для хранения значений,
// а структура Metric используется для API и сериализации.
package models

import "errors"

// Gauge — тип для хранения значения метрики типа gauge (вещественное число).
type Gauge float64

// Counter — тип для хранения значения метрики типа counter (целое число с накоплением).
type Counter int64

// MetricType — строковый тип, обозначающий вид метрики.
type MetricType string

const (
	// TypeCounter — тип метрики «счётчик», значение которого только растёт.
	TypeCounter MetricType = "counter"

	// TypeGauge — тип метрики «измеритель», значение которого может произвольно меняться.
	TypeGauge MetricType = "gauge"
)

// ValidMetricTypes содержит множество допустимых типов метрик.
// Используется для быстрой проверки корректности типа.
var ValidMetricTypes = map[MetricType]bool{
	TypeCounter: true,
	TypeGauge:   true,
}

var (
	// ErrInvalidType возвращается, если тип метрики не поддерживается.
	ErrInvalidType = errors.New("unsupported metric type")

	// ErrNoName возвращается, если имя метрики не указано.
	ErrNoName = errors.New("metric name not provided")

	// ErrInvalidValue возвращается, если у метрики типа gauge не задано значение Value.
	ErrInvalidValue = errors.New("missing value for gauge metric")

	// ErrInvalidDelta возвращается, если у метрики типа counter не задано значение Delta.
	ErrInvalidDelta = errors.New("missing delta for counter metric")
)

// Metric — универсальная структура метрики, используемая в API и хранилище.
//
// Delta и Value объявлены через указатели, чтобы различать значение "0"
// от незаданного значения и не включать пустые поля в JSON-сериализацию.
//
// Пример gauge-метрики:
//
//	m := models.NewMetricGauge("cpu_usage", 72.5)
//
// Пример counter-метрики:
//
//	m := models.NewMetricCounter("requests_total", 42)
type Metric struct {
	// ID — уникальное имя метрики.
	ID string `json:"id"`

	// MType — тип метрики: "gauge" или "counter".
	MType MetricType `json:"type"`

	// Delta — значение счётчика (только для counter).
	Delta *Counter `json:"delta,omitempty"`

	// Value — значение измерителя (только для gauge).
	Value *Gauge `json:"value,omitempty"`

	// Hash — опциональная подпись для верификации значения метрики.
	Hash string `json:"hash,omitempty"`
}

// NewMetricGauge создаёт новую метрику типа gauge с заданным именем и значением.
//
//	m := models.NewMetricGauge("temperature", 36.6)
func NewMetricGauge(id string, value Gauge) *Metric {
	return &Metric{
		ID:    id,
		MType: TypeGauge,
		Value: &value,
	}
}

// NewMetricCounter создаёт новую метрику типа counter с заданным именем и значением дельты.
//
//	m := models.NewMetricCounter("http_requests", 1)
func NewMetricCounter(id string, delta Counter) *Metric {
	return &Metric{
		ID:    id,
		MType: TypeCounter,
		Delta: &delta,
	}
}

// ValidateMeta проверяет, что у метрики задано имя и корректный тип.
// Возвращает ErrNoName или ErrInvalidType при нарушении.
func (met *Metric) ValidateMeta() error {
	if met.ID == "" {
		return ErrNoName
	}

	if !IsValidMetricType(met.MType) {
		return ErrInvalidType
	}

	return nil
}

// ValidateValue проверяет, что у метрики задано значение, соответствующее её типу:
// для gauge — поле Value, для counter — поле Delta.
// Возвращает ErrInvalidValue, ErrInvalidDelta или ErrInvalidType.
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

// IsValidMetricType возвращает true, если переданный тип метрики поддерживается сервисом.
func IsValidMetricType(metType MetricType) bool {
	return ValidMetricTypes[metType]
}
