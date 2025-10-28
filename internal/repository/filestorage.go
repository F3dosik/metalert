package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/F3dosik/metalert.git/pkg/models"
)

type FileMetricsStorage struct {
	*MemMetricsStorage
	fileName string
	tmpFile  *os.File
}

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

func (f *FileMetricsStorage) Save(ctx context.Context) error {
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

func (f *FileMetricsStorage) Close(ctx context.Context) error {
	if err := f.Save(ctx); err != nil {
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
