package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileAuditObserver_Notify_Success(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "audit-*.log")
	require.NoError(t, err)
	f.Close()

	observer := NewFileAuditObserver(f.Name())
	event := AuditEvent{
		Ts:        12345678,
		Metrics:   []string{"Alloc", "Frees"},
		IPAddress: "192.168.0.42",
	}

	err = observer.Notify(event)
	require.NoError(t, err)

	file, err := os.Open(f.Name())
	require.NoError(t, err)
	defer file.Close()

	var received AuditEvent
	require.NoError(t, json.NewDecoder(file).Decode(&received))
	assert.Equal(t, event, received)
}

func TestFileAuditObserver_Notify_Append(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "audit-*.log")
	require.NoError(t, err)
	f.Close()

	observer := NewFileAuditObserver(f.Name())
	events := []AuditEvent{
		{Ts: 1, Metrics: []string{"Alloc"}, IPAddress: "10.0.0.1"},
		{Ts: 2, Metrics: []string{"Frees"}, IPAddress: "10.0.0.2"},
		{Ts: 3, Metrics: []string{"Sys"}, IPAddress: "10.0.0.3"},
	}

	for _, e := range events {
		require.NoError(t, observer.Notify(e))
	}

	file, err := os.Open(f.Name())
	require.NoError(t, err)
	defer file.Close()

	var received []AuditEvent
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e AuditEvent
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &e))
		received = append(received, e)
	}
	require.NoError(t, scanner.Err())

	assert.Equal(t, events, received)
}

func TestFileAuditObserver_Notify_InvalidPath(t *testing.T) {
	observer := NewFileAuditObserver("/nonexistent/path/audit.log")
	err := observer.Notify(AuditEvent{Ts: 1, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"})

	assert.Error(t, err)
	assert.ErrorContains(t, err, "open file")
}
