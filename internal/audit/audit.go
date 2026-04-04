// Package audit реализует систему аудита событий обновления метрик
// на основе паттерна «Наблюдатель» (Observer).
//
// При каждом обновлении метрики хендлер публикует [AuditEvent] через [AuditDispatcher].
// Диспетчер асинхронно уведомляет всех зарегистрированных наблюдателей ([AuditObserver]).
//
// Доступные реализации наблюдателей:
//   - [FileAuditObserver] — записывает события в JSON-файл
//   - [URLAuditObserver] — отправляет события на HTTP-эндпоинт
//
// Пример использования:
//
//	dispatcher := audit.NewAuditDispatcher(logger)
//	dispatcher.Register(audit.NewFileAuditObserver("/var/log/audit.jsonl"))
//	dispatcher.Register(audit.NewURLAuditObserver("https://audit.example.com/events"))
//
//	// при обновлении метрики:
//	dispatcher.Publish(audit.AuditEvent{
//	    Ts:        time.Now().Unix(),
//	    Metrics:   []string{"cpu_usage"},
//	    IPAddress: "192.168.1.1",
//	})
//
//	// дождаться завершения всех горутин перед выходом:
//	dispatcher.Wait()
package audit

//go:generate mockery

import (
	"io"
	"sync"

	"go.uber.org/zap"
)

// AuditEvent описывает одно событие аудита — факт обновления метрик агентом.
type AuditEvent struct {
	// Ts — Unix-время события в секундах.
	Ts int64 `json:"ts"`

	// Metrics — список имён метрик, обновлённых в рамках одного запроса.
	Metrics []string `json:"metrics"`

	// IPAddress — IP-адрес клиента, приславшего обновление.
	IPAddress string `json:"ip_address"`
}

// AuditObserver — интерфейс наблюдателя, получающего аудит-события.
//
// Реализуется типами [FileAuditObserver] и [URLAuditObserver].
// Метод Notify вызывается синхронно внутри горутины диспетчера,
// поэтому реализация должна быть потокобезопасной.
type AuditObserver interface {
	// Notify обрабатывает аудит-событие. Возвращает ошибку при сбое,
	// которая логируется диспетчером, но не прерывает уведомление остальных наблюдателей.
	Notify(event AuditEvent) error
}

// AuditDispatcher рассылает аудит-события всем зарегистрированным наблюдателям.
//
// Потокобезопасен: [Register] и [Unregister] можно вызывать из любой горутины.
// После вызова [Wait] диспетчер переходит в остановленное состояние:
// последующие вызовы [Publish] молча игнорируются — паники не будет.
// Для ожидания завершения всех горутин перед остановкой сервера используйте [AuditDispatcher.Wait].
type AuditDispatcher struct {
	mu        sync.RWMutex
	observers []AuditObserver
	stopped   bool

	logger *zap.SugaredLogger
	wg     sync.WaitGroup
}

// NewAuditDispatcher создаёт новый диспетчер без наблюдателей.
// Наблюдатели добавляются через [AuditDispatcher.Register].
func NewAuditDispatcher(logger *zap.SugaredLogger) *AuditDispatcher {
	return &AuditDispatcher{logger: logger}
}

// Register добавляет наблюдателя в список рассылки.
// Потокобезопасен, можно вызывать параллельно с Publish.
func (d *AuditDispatcher) Register(o AuditObserver) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.observers = append(d.observers, o)
}

// Unregister удаляет первого наблюдателя, совпадающего с o по указателю.
// Если наблюдатель не найден — вызов является no-op.
// Потокобезопасен, можно вызывать параллельно с Publish.
func (d *AuditDispatcher) Unregister(o AuditObserver) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, obs := range d.observers {
		if obs == o {
			last := len(d.observers) - 1
			d.observers[i] = d.observers[last]
			d.observers[last] = nil
			d.observers = d.observers[:last]
			return
		}
	}
}

// publish синхронно уведомляет всех наблюдателей под RLock.
func (d *AuditDispatcher) publish(event AuditEvent) {
	d.mu.RLock()
	observers := make([]AuditObserver, len(d.observers))
	copy(observers, d.observers)
	d.mu.RUnlock()

	for _, obs := range observers {
		if err := obs.Notify(event); err != nil {
			d.logger.Warnw("audit notify error", "error", err)
		}
	}
}

// Publish асинхронно рассылает событие всем наблюдателям.
// Если диспетчер уже остановлен (был вызван Wait), событие молча игнорируется.
func (d *AuditDispatcher) Publish(event AuditEvent) {
	d.mu.Lock()
	if d.stopped {
		d.mu.Unlock()
		return
	}
	d.wg.Add(1)
	d.mu.Unlock()

	go func() {
		defer d.wg.Done()
		d.publish(event)
	}()
}

// Wait блокирует выполнение до завершения всех запущенных горутин Publish.
// Следует вызывать при graceful shutdown сервера.
func (d *AuditDispatcher) Wait() {
	d.mu.Lock()
	d.stopped = true
	d.mu.Unlock()

	d.wg.Wait()
}

func (d *AuditDispatcher) Close() {
	d.Wait()

	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, obs := range d.observers {
		if closer, ok := obs.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				d.logger.Warnw("audit observer close error", "error", err)
			}
		}
	}
}
