package audit

import (
	"sync"

	"go.uber.org/zap"
)

type AuditEvent struct {
	Ts        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type AuditObserver interface {
	Notify(event AuditEvent) error
}

type AuditDispatcher struct {
	observers []AuditObserver
	logger    *zap.SugaredLogger
	wg        sync.WaitGroup
}

func NewAuditDispatcher(logger *zap.SugaredLogger) *AuditDispatcher {
	return &AuditDispatcher{logger: logger}
}

func (d *AuditDispatcher) Register(o AuditObserver) {
	d.observers = append(d.observers, o)
}

func (d *AuditDispatcher) publish(event AuditEvent) {
	for _, obs := range d.observers {
		if err := obs.Notify(event); err != nil {
			d.logger.Warnw("audit notify error", "error", err)
		}
	}
}

func (d *AuditDispatcher) Publish(event AuditEvent) {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.publish(event)
	}()
}

func (d *AuditDispatcher) Wait() {
	d.wg.Wait()
}
