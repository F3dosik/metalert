package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/F3dosik/metalert.git/pkg/models"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func runMigrations(dsn string) error {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

type DBMetricsStorage struct {
	db *sql.DB
}

func NewDBMetricStorage(dsn string) (*DBMetricsStorage, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &DBMetricsStorage{
		db: db,
	}, nil
}

func (d *DBMetricsStorage) SetGauge(ctx context.Context, name string, value models.Gauge) error {
	_, err := d.db.ExecContext(ctx, `
		INSERT INTO metrics (id,type,value)
		VALUES ($1, 'gauge', $2)
		ON CONFLICT (id) DO UPDATE
		SET value = EXCLUDED.value;
	`, name, value)
	if err != nil {
		return fmt.Errorf("failed to upsert gauge: %w", err)
	}

	return nil
}

func (d *DBMetricsStorage) GetGauge(ctx context.Context, name string) (models.Gauge, error) {
	var value models.Gauge
	row := d.db.QueryRowContext(ctx, `
		SELECT value FROM metrics
		WHERE id = $1
	`, name)

	err := row.Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("gauge %q not found", name)
		}
		return 0, fmt.Errorf("failed to get gauge %q: %w", name, err)
	}

	return value, nil
}

func (d *DBMetricsStorage) AddCounter(ctx context.Context, name string, value models.Counter) error {
	_, err := d.db.ExecContext(ctx, `
		INSERT INTO metrics (id,type,delta)
		VALUES ($1, 'counter', $2)
		ON CONFLICT (id) DO UPDATE
		SET delta = metrics.delta + EXCLUDED.delta;
	`, name, value)
	if err != nil {
		return fmt.Errorf("failed to upsert counter: %w", err)
	}

	return nil
}

func (d *DBMetricsStorage) GetCounter(ctx context.Context, name string) (models.Counter, error) {
	var delta models.Counter
	row := d.db.QueryRowContext(ctx, `
		SELECT delta FROM metrics
		WHERE id = $1
	`, name)
	err := row.Scan(&delta)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("counter %q not found", name)
		}
		return 0, fmt.Errorf("failed to get counter %q: %w", name, err)
	}

	return delta, nil
}

func (d *DBMetricsStorage) GetAllMetrics(ctx context.Context) ([]models.Metric, error) {
	metrics := make([]models.Metric, 0)
	rows, err := d.db.QueryContext(ctx, `
		SELECT id,type,value,delta FROM metrics
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var mtype string
		var value sql.NullFloat64
		var delta sql.NullInt64

		if err := rows.Scan(&id, &mtype, &value, &delta); err != nil {
			return nil, err
		}

		metric := models.Metric{
			ID:    id,
			MType: models.MetricType(mtype),
		}
		if mtype == string(models.TypeGauge) && value.Valid {
			v := models.Gauge(value.Float64)
			metric.Value = &v
		}
		if mtype == string(models.TypeCounter) && delta.Valid {
			d := models.Counter(delta.Int64)
			metric.Delta = &d
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, err

	}

	return metrics, nil
}

func (s *DBMetricsStorage) UpdateMetricTx(ctx context.Context, metrics []models.Metric) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `
		INSERT INTO metrics (id, type, value, delta)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE 
		SET 
			value = COALESCE(EXCLUDED.value, metrics.value),
			delta = COALESCE(metrics.delta + EXCLUDED.delta, metrics.delta);
	`
	stmt, err := tx.PrepareContext(ctx, query)

	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, metric := range metrics {
		switch metric.MType {
		case models.TypeGauge:
			_, err := stmt.ExecContext(ctx, metric.ID, "gauge", metric.Value, nil)
			if err != nil {
				return fmt.Errorf("failed to update gauge %s: %w", metric.ID, err)
			}
		case models.TypeCounter:
			_, err := stmt.ExecContext(ctx, metric.ID, "counter", nil, metric.Delta)
			if err != nil {
				return fmt.Errorf("failed to update counter %s: %w", metric.ID, err)
			}
		}
	}
	return tx.Commit()
}

func (d *DBMetricsStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := d.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
