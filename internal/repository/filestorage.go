package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/F3dosik/metalert/pkg/models"
)

// FileMetricsStorage — персистентное хранилище метрик на основе [MemMetricsStorage].
//
// Все операции чтения и записи делегируются встроенному MemMetricsStorage.
// Метод [FileMetricsStorage.Save] атомарно сбрасывает текущее состояние на диск
// через временный файл с последующим переименованием, что защищает от частичной записи.
//
// Пример создания с восстановлением данных из файла:
//
//	storage, err := repository.NewFileMetricsStorage("/var/lib/metalert/metrics.json", true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer storage.Close(ctx)
type FileMetricsStorage struct {
	*MemMetricsStorage
	fileName string
	tmpFile  *os.File
}

// NewFileMetricsStorage создаёт FileMetricsStorage, связанное с файлом fileName.
//
// Если restore=true, при старте загружает ранее сохранённые метрики из файла.
// Автоматически создаёт все необходимые директории.
// Для освобождения ресурсов необходимо вызвать [FileMetricsStorage.Close].
func NewFileMetricsStorage(fileName string, restore bool) (*FileMetricsStorage, error) {
	memStore := NewMemMetricsStorage()

	absPath, err := filepath.Abs(fileName)
	if err != nil {
		absPath = fileName
	}

	dir := filepath.Dir(absPath)
	if err = os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("ошибка создания каталога %s: %w", dir, err)
	}

	tmpFile, err := os.OpenFile(absPath+".tmp", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	fs := &FileMetricsStorage{
		MemMetricsStorage: memStore,
		fileName:          absPath,
		tmpFile:           tmpFile,
	}

	if restore {
		fs.load(context.Background())
	}

	return fs, nil
}

// Save сохраняет все текущие метрики в файл атомарно:
// сначала записывает во временный файл, затем переименовывает в целевой.
//
// Реализует интерфейс [Savable], что позволяет хендлерам вызывать Save
// асинхронно после каждого обновления метрики.
func (f *FileMetricsStorage) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	metrics, _ := f.GetAllMetrics(ctx)

	data, err := json.Marshal(&metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics: %w", err)
	}

	if err := f.tmpFile.Truncate(0); err != nil {
		return err
	}
	if _, err := f.tmpFile.Seek(0, 0); err != nil {
		return err
	}

	if _, err := f.tmpFile.Write(data); err != nil {
		return fmt.Errorf("write tmp file: %w", err)
	}

	return nil
}

// load восстанавливает метрики из файла fileName в память.
// Вызывается автоматически из NewFileMetricsStorage при restore=true.
func (f *FileMetricsStorage) load(ctx context.Context) error {
	data, err := os.ReadFile(f.fileName)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var metrics []models.Metric
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("unmarshal metrics: %w", err)
	}

	for _, metric := range metrics {
		switch metric.MType {
		case models.TypeGauge:
			f.SetGauge(ctx, metric.ID, *metric.Value)
		case models.TypeCounter:
			f.SetCounter(ctx, metric.ID, *metric.Delta)
		}
	}

	return nil
}

// Close сохраняет метрики, синхронизирует и закрывает временный файл,
// после чего атомарно переименовывает его в целевой fileName.
//
// Следует вызывать при завершении работы сервера (например, через defer).
func (f *FileMetricsStorage) Close() error {
	if err := f.Save(); err != nil {
		return fmt.Errorf("save before close: %w", err)
	}

	if err := f.tmpFile.Sync(); err != nil {
		return fmt.Errorf("sync tmp file: %w", err)
	}

	if err := f.tmpFile.Close(); err != nil {
		return fmt.Errorf("close tmp file: %w", err)
	}

	if err := os.Rename(f.fileName+".tmp", f.fileName); err != nil {
		return fmt.Errorf("rename tmp file: %w", err)
	}

	return nil
}
