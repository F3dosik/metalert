package audit_test

import (
	"fmt"
	"testing"

	"github.com/F3dosik/metalert.git/internal/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	mocks "github.com/F3dosik/metalert.git/internal/audit/mocks"
)

func newTestDispatcher() *audit.AuditDispatcher {
	return audit.NewAuditDispatcher(zap.NewNop().Sugar())
}

func TestAuditDispatcher_Publish_AllObserversReceiveEvent(t *testing.T) {
	d := newTestDispatcher()

	o1 := mocks.NewMockAuditObserver(t)
	o2 := mocks.NewMockAuditObserver(t)

	event := audit.AuditEvent{Ts: 99999, Metrics: []string{"Alloc", "Frees"}, IPAddress: "10.0.0.1"}

	o1.EXPECT().Notify(event).Return(nil).Once()
	o2.EXPECT().Notify(event).Return(nil).Once()

	d.Register(o1)
	d.Register(o2)
	d.Publish(event)
	d.Wait()
}

func TestAuditDispatcher_Publish_ObserverError_DoesNotStopOthers(t *testing.T) {
	d := newTestDispatcher()

	failing := mocks.NewMockAuditObserver(t)
	succeeding := mocks.NewMockAuditObserver(t)

	event := audit.AuditEvent{Ts: 1, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"}

	failing.EXPECT().Notify(event).Return(fmt.Errorf("some error")).Once()
	succeeding.EXPECT().Notify(event).Return(nil).Once()

	d.Register(failing)
	d.Register(succeeding)
	d.Publish(event)
	d.Wait()
}

func TestAuditDispatcher_Publish_NoObservers(t *testing.T) {
	d := newTestDispatcher()
	assert.NotPanics(t, func() {
		d.Publish(audit.AuditEvent{Ts: 1, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"})
		d.Wait()
	})
}

func TestAuditDispatcher_Unregister_ObserverNotNotified(t *testing.T) {
	d := newTestDispatcher()

	obs := mocks.NewMockAuditObserver(t)
	obs.EXPECT().Notify(mock.Anything).Return(nil).Maybe()

	d.Register(obs)
	d.Unregister(obs)
	d.Publish(audit.AuditEvent{Ts: 2})
	d.Wait()
}

func TestAuditDispatcher_Publish_AfterWait_IsIgnored(t *testing.T) {
	d := newTestDispatcher()

	obs := mocks.NewMockAuditObserver(t)
	obs.EXPECT().Notify(mock.Anything).Return(nil).Maybe()

	d.Register(obs)
	d.Wait()
	assert.NotPanics(t, func() {
		d.Publish(audit.AuditEvent{Ts: 3})
	})
}
