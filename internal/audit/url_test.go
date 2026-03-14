package audit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLAuditObserver_Notify_Success(t *testing.T) {
	var received AuditEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &received))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := NewURLAuditObserver(server.URL)
	event := AuditEvent{
		Ts:        12345678,
		Metrics:   []string{"Alloc", "Frees"},
		IPAddress: "192.168.0.42",
	}

	err := observer.Notify(event)

	require.NoError(t, err)
	assert.Equal(t, event, received)
}

func TestURLAuditObserver_Notify_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	observer := NewURLAuditObserver(server.URL)
	err := observer.Notify(AuditEvent{Ts: 1, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"})

	require.Error(t, err)
	assert.ErrorContains(t, err, "500")
}

func TestURLAuditObserver_Notify_Unreachable(t *testing.T) {
	observer := NewURLAuditObserver("http://127.0.0.1:1")
	err := observer.Notify(AuditEvent{Ts: 1, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"})

	assert.Error(t, err)
}
