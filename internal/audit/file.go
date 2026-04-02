package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// FileAuditObserver — наблюдатель, записывающий аудит-события в файл в формате JSON Lines.
//
// Файл открывается один раз при создании наблюдателя и закрывается явным вызовом [FileAuditObserver.Close].
//
// Пример:
//
//	observer, err := audit.NewFileAuditObserver("/var/log/metrics-audit.jsonl")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer observer.Close()
//	dispatcher.Register(observer)
type FileAuditObserver struct {
	mu      sync.Mutex
	file    *os.File
	encoder *json.Encoder
}

// NewFileAuditObserver создаёт наблюдателя и открывает файл по пути path.
// Файл создаётся автоматически если не существует.
// Вызывающий обязан вызвать [FileAuditObserver.Close] по завершении работы.
func NewFileAuditObserver(path string) (*FileAuditObserver, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}
	return &FileAuditObserver{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// Notify записывает событие в файл в формате JSON.
func (o *FileAuditObserver) Notify(event AuditEvent) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := o.encoder.Encode(event); err != nil {
		return fmt.Errorf("encode event: %w", err)
	}
	return nil
}

// Close закрывает файл. Должен вызываться при завершении программы.
// Повторный вызов вернёт ошибку от os.File.Close.
func (o *FileAuditObserver) Close() error {
	return o.file.Close()
}
