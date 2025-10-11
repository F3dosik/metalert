// Package service реализует бизнес-логику сервера метрик.
// Здесь функции проверки и преобразования значений метрик,
// а также взаимодействие между хранилищем и API.
package service

import (
	"strconv"

	"github.com/F3dosik/metalert.git/pkg/models"
)

func CheckAndParseValue(metType models.MetricType, metName, metValue string) (any, error) {
	// log.Printf("CheckAndParseValue: type=%s, name=%s, value=%s", metType, metName, metValue)

	if err := ValidateMetricType(metType); err != nil {
		// log.Printf("ValidateMetricType failed: %v", err)
		return nil, err
	}

	if err := validateMetricName(metName); err != nil {
		// log.Printf("validateMetricName failed: %v", err)
		return nil, err
	}

	switch metType {
	case models.TypeGauge:
		value, err := parseGaugeValue(metValue)
		// log.Printf("parseGaugeValue result: %v, error: %v", value, err)
		return value, err
	case models.TypeCounter:
		value, err := parseCounterValue(metValue)
		// log.Printf("parseCounterValue result: %v, error: %v", value, err)
		return value, err
	default:
		// log.Printf("Unknown metric type: %s", metType)
		return nil, models.ErrInvalidType
	}
}

func ValidateMetricType(metType models.MetricType) error {
	// log.Printf("ValidateMetricType: %s", metType)
	if !models.IsValidMetricType(metType) {
		return models.ErrInvalidType
	}
	return nil
}

func validateMetricName(metName string) error {
	// log.Printf("validateMetricName: %s", metName)
	if metName == "" {
		return models.ErrNoName
	}
	return nil
}

func parseGaugeValue(metValue string) (models.Gauge, error) {
	// log.Printf("parseCounterValue: %s", metValue)
	f, err := strconv.ParseFloat(metValue, 64)
	if err != nil {
		return 0, models.ErrInvalidValue
	}
	return models.Gauge(f), nil
}

func parseCounterValue(metValue string) (models.Counter, error) {
	// log.Printf("parseGaugeValue: %s", metValue)
	i, err := strconv.ParseInt(metValue, 10, 64)
	if err != nil {
		return 0, models.ErrInvalidDelta
	}
	return models.Counter(i), nil
}
