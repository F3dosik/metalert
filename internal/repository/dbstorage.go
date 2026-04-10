package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/F3dosik/metalert/internal/pgerrors"
	"github.com/F3dosik/metalert/pkg/models"
)

// runMigrations применяет SQL-миграции из директории migrations/ к базе данных по dsn.
// Использует golang-migrate. Если миграции уже применены, возвращает nil.
func runMigrations(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	err = withRetry(context.Background(), func() error {
		e := m.Up()
		if errors.Is(e, migrate.ErrNoChange) {
			return nil
		}
		return e
	})

	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

// DBMetricsStorage — хранилище метрик на основе PostgreSQL.
//
// Поддерживает операции upsert для gauge и counter, пакетное обновление
// в транзакции через [DBMetricsStorage.UpdateMetricTx], а также
// автоматический retry при временно недоступной БД.
//
// Создаётся через [NewDBMetricStorage], который автоматически применяет миграции.
type DBMetricsStorage struct {
	db *sql.DB
}

// NewDBMetricStorage создаёт DBMetricsStorage: применяет миграции, открывает
// соединение с PostgreSQL и проверяет его через Ping.
//
// dsn — строка подключения в формате PostgreSQL DSN:
//
//	"postgres://user:password@localhost:5432/metrics?sslmode=disable"
func NewDBMetricStorage(dsn string) (*DBMetricsStorage, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	storage := &DBMetricsStorage{db: db}
	if err = storage.Ping(); err != nil {
		return nil, err
	}

	return storage, nil
}

// SetGauge устанавливает значение gauge-метрики в БД.
// При конфликте по id перезаписывает существующее значение (upsert).
func (d *DBMetricsStorage) SetGauge(ctx context.Context, name string, value models.Gauge) error {
	err := withRetry(ctx, func() error {
		_, err := d.db.ExecContext(ctx, `
			INSERT INTO metrics (id,type,value)
			VALUES ($1, 'gauge', $2)
			ON CONFLICT (id) DO UPDATE
			SET value = EXCLUDED.value;
		`, name, value)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to upsert gauge: %w", err)
	}

	return nil
}

// GetGauge возвращает текущее значение gauge-метрики из БД.
// Возвращает ошибку, если метрика не найдена.
func (d *DBMetricsStorage) GetGauge(ctx context.Context, name string) (models.Gauge, error) {
	var value models.Gauge

	row := d.db.QueryRowContext(ctx, `
		SELECT value FROM metrics
		WHERE id = $1
	`, name)

	err := withRetry(ctx, func() error {
		return row.Scan(&value)
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("gauge %q not found", name)
		}
		return 0, fmt.Errorf("failed to get gauge %q: %w", name, err)
	}

	return value, nil
}

// AddCounter прибавляет value к текущему значению counter-метрики в БД.
// Если метрика не существует, создаётся с переданным значением (upsert).
func (d *DBMetricsStorage) AddCounter(ctx context.Context, name string, value models.Counter) error {
	err := withRetry(ctx, func() error {
		_, err := d.db.ExecContext(ctx, `
			INSERT INTO metrics (id,type,delta)
			VALUES ($1, 'counter', $2)
			ON CONFLICT (id) DO UPDATE
			SET delta = metrics.delta + EXCLUDED.delta;
		`, name, value)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to upsert counter: %w", err)
	}

	return nil
}

// GetCounter возвращает текущее значение counter-метрики из БД.
// Возвращает ошибку, если метрика не найдена.
func (d *DBMetricsStorage) GetCounter(ctx context.Context, name string) (models.Counter, error) {
	var delta models.Counter

	row := d.db.QueryRowContext(ctx, `
		SELECT delta FROM metrics
		WHERE id = $1
	`, name)

	err := withRetry(ctx, func() error {
		return row.Scan(&delta)
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("counter %q not found", name)
		}
		return 0, fmt.Errorf("failed to get counter %q: %w", name, err)
	}

	return delta, nil
}

// GetAllMetrics возвращает все метрики из таблицы metrics.
func (d *DBMetricsStorage) GetAllMetrics(ctx context.Context) ([]models.Metric, error) {
	var metrics []models.Metric

	err := withRetry(ctx, func() error {
		rows, err := d.db.QueryContext(ctx, `
			SELECT id,type,value,delta FROM metrics
		`)
		if err != nil {
			return err
		}
		defer rows.Close()

		var tmpMetrics []models.Metric
		for rows.Next() {
			var id, mtype string
			var value sql.NullFloat64
			var delta sql.NullInt64

			if err := rows.Scan(&id, &mtype, &value, &delta); err != nil {
				return err
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

			tmpMetrics = append(tmpMetrics, metric)
		}

		if err := rows.Err(); err != nil {
			return err
		}

		metrics = tmpMetrics
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get all metrics: %w", err)
	}

	return metrics, nil
}

// UpdateMetricTx обновляет несколько метрик в одной транзакции.
//
// Для gauge выполняет upsert значения Value.
// Для counter прибавляет Delta к существующему значению.
// При ошибке транзакция откатывается.
// Используется хендлером [handler.UpdatesJSONHandler] для пакетного обновления.
// UpdateMany обновляет набор метрик в одной транзакции.
func (d *DBMetricsStorage) UpdateMany(ctx context.Context, metrics []models.Metric) error {
	return d.UpdateMetricTx(ctx, metrics)
}

func (d *DBMetricsStorage) UpdateMetricTx(ctx context.Context, metrics []models.Metric) error {
	return withRetry(ctx, func() error {
		tx, err := d.db.BeginTx(ctx, nil)
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
	})
}

// Ping проверяет соединение с базой данных с таймаутом 10 секунд.
// Используется хендлером [handler.PingDB] для проверки доступности БД.
func (d *DBMetricsStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return withRetry(ctx, func() error {
		return d.db.PingContext(ctx)
	})
}

// withRetry выполняет операцию op с повторными попытками при retriable-ошибках PostgreSQL.
//
// Стратегия задержек: 1с, 3с, 5с (итого до 4 попыток).
// Прекращает попытки при отмене контекста или не-retriable ошибке.
func withRetry(ctx context.Context, op func() error) error {
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	maxAttempts := len(delays) + 1

	var err error
	for i := 0; i < maxAttempts; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err = op()
		if err == nil {
			return nil
		}

		if !pgerrors.IsRetriable(err) || i == len(delays) {
			return fmt.Errorf("operation failed after %d attempt(s): %w", i+1, err)
		}

		time.Sleep(delays[i])
	}

	return fmt.Errorf("operation failed after %d attempt(s): %w", maxAttempts, err)
}
