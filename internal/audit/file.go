package audit

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileAuditObserver struct {
	path string
}

func NewFileAuditObserver(path string) *FileAuditObserver {
	return &FileAuditObserver{path: path}
}

func (o *FileAuditObserver) Notify(event AuditEvent) error {
	file, err := os.OpenFile(o.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(event); err != nil {
		return fmt.Errorf("encode event: %w", err)
	}
	return nil
}
