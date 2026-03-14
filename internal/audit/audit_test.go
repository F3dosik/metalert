package audit

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockObserver struct {
	mu     sync.Mutex
	events []AuditEvent
	err    error
}

func (m *mockObserver) Notify(event AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return m.err
}

func (m *mockObserver) Events() []AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events
}

func newTestDispatcher() *AuditDispatcher {
	return &AuditDispatcher{logger: zap.NewNop().Sugar()}
}

func TestAuditDispatcher_Publish_AllObserversReceiveEvent(t *testing.T) {
	d := newTestDispatcher()
	o1 := &mockObserver{}
	o2 := &mockObserver{}
	d.Register(o1)
	d.Register(o2)

	event := AuditEvent{Ts: 99999, Metrics: []string{"Alloc", "Frees"}, IPAddress: "10.0.0.1"}
	d.Publish(event)
	d.Wait()

	require.Len(t, o1.Events(), 1)
	assert.Equal(t, event, o1.Events()[0])
	require.Len(t, o2.Events(), 1)
	assert.Equal(t, event, o2.Events()[0])
}

func TestAuditDispatcher_Publish_ObserverError_DoesNotStopOthers(t *testing.T) {
	d := newTestDispatcher()
	failing := &mockObserver{err: fmt.Errorf("some error")}
	succeeding := &mockObserver{}
	d.Register(failing)
	d.Register(succeeding)

	d.Publish(AuditEvent{Ts: 1, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"})
	d.Wait()

	assert.Len(t, succeeding.Events(), 1)
}

func TestAuditDispatcher_Publish_NoObservers(t *testing.T) {
	d := &AuditDispatcher{}
	assert.NotPanics(t, func() {
		d.Publish(AuditEvent{Ts: 1, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"})
	})
}
