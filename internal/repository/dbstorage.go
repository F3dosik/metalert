package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/F3dosik/metalert.git/pkg/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBMetricStorage struct {
	db *sql.DB
}

func NewDBMetricStorage(dsn string) (*DBMetricStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &DBMetricStorage{
		db: db,
	}, nil
}

func (d *DBMetricStorage) SetGauge(name string, value models.Gauge)   {}
func (d *DBMetricStorage) GetGauge(name string) (models.Gauge, error) { return models.Gauge(0), nil }

func (d *DBMetricStorage) SetCounter(name string, value models.Counter) {}
func (d *DBMetricStorage) AddCounter(name string, value models.Counter) {}
func (d *DBMetricStorage) GetCounter(name string) (models.Counter, error) {
	return models.Counter(0), nil
}

func (d *DBMetricStorage) GetAllMetrics() []models.Metric { return nil }

func (d *DBMetricStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := d.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
