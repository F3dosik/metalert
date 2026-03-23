// Package service содержит бизнес-логику обработки метрик:
// валидацию входных данных и обновление значений в хранилище.
package service

import (
	"log"
	"strconv"

	"github.com/F3dosik/metalert/pkg/models"
)

// CheckAndParseValue валидирует тип и имя метрики, затем парсит строковое значение
// в соответствующий Go-тип: [models.Gauge] или [models.Counter].
//
// Используется хендлером [handler.UpdateHandler] для обработки URL-параметров.
//
// Возвращаемые ошибки:
//   - [models.ErrInvalidType] — неизвестный тип метрики
//   - [models.ErrNoName] — пустое имя метрики
//   - [models.ErrInvalidValue] — значение не является числом с плавающей точкой (для gauge)
//   - [models.ErrInvalidDelta] — значение не является целым числом (для counter)
func CheckAndParseValue(metType models.MetricType, metName, metValue string) (any, error) {
	if err := ValidateMetricType(metType); err != nil {
		return nil, err
	}

	if err := validateMetricName(metName); err != nil {
		return nil, err
	}

	switch metType {
	case models.TypeGauge:
		value, err := parseGaugeValue(metValue)
		return value, err
	case models.TypeCounter:
		value, err := parseCounterValue(metValue)
		return value, err
	default:
		log.Printf("Unknown metric type: %s", metType)
		return nil, models.ErrInvalidType
	}
}

// ValidateMetricType проверяет, что metType входит в множество допустимых типов.
// Возвращает [models.ErrInvalidType], если тип не поддерживается.
func ValidateMetricType(metType models.MetricType) error {
	log.Printf("ValidateMetricType: %s", metType)
	if !models.IsValidMetricType(metType) {
		return models.ErrInvalidType
	}
	return nil
}

// validateMetricName проверяет, что имя метрики не пустое.
// Возвращает [models.ErrNoName], если metName == "".
func validateMetricName(metName string) error {
	if metName == "" {
		return models.ErrNoName
	}
	return nil
}

// parseGaugeValue разбирает строку metValue как число с плавающей точкой (float64).
// Возвращает [models.ErrInvalidValue] при ошибке разбора.
func parseGaugeValue(metValue string) (models.Gauge, error) {
	f, err := strconv.ParseFloat(metValue, 64)
	if err != nil {
		return 0, models.ErrInvalidValue
	}
	return models.Gauge(f), nil
}

// parseCounterValue разбирает строку metValue как целое знаковое число (int64).
// Возвращает [models.ErrInvalidDelta] при ошибке разбора.
func parseCounterValue(metValue string) (models.Counter, error) {
	i, err := strconv.ParseInt(metValue, 10, 64)
	if err != nil {
		return 0, models.ErrInvalidDelta
	}
	return models.Counter(i), nil
}
