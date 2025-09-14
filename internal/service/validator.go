// Package service реализует бизнес-логику сервера метрик.
// Здесь функции проверки и преобразования значений метрик,
// а также взаимодействие между хранилищем и API.
package service

import (
	"errors"
	"fmt"
	"strconv"

	models "github.com/F3dosik/metalert.git/internal/model"
)

var (
	ErrInvalidType = errors.New("error: unknown metric type")
	ErrNoName      = errors.New("error: metric name not provided")
	ErrInvalidVal  = errors.New("error: invalid value for")
)

func CheckURL(mT, mN, mV string) (any, error) {
	switch mT {
	case models.GaugeType:
		if err := checkName(mN); err != nil {
			return nil, err
		}
		return checkGauge(mV)
	case models.CounterType:
		if err := checkName(mN); err != nil {
			return nil, err
		}
		return checkCounter(mV)
	default:
		return nil, ErrInvalidType
	}
}

func checkName(mN string) error {
	if mN == "" {
		return ErrNoName
	}
	return nil
}

func checkGauge(s string) (models.Gauge, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidVal, s)
	}
	return models.Gauge(f), nil
}

func checkCounter(s string) (models.Counter, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidVal, s)
	}
	return models.Counter(i), nil
}
