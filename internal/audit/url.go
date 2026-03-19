package audit

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// URLAuditObserver — наблюдатель, отправляющий аудит-события на HTTP-эндпоинт через POST-запрос.
//
// Тело запроса — JSON-сериализованный [AuditEvent].
// Таймаут запроса — 5 секунд.
//
// Пример:
//
//	observer := audit.NewURLAuditObserver("https://audit.example.com/events")
//	dispatcher.Register(observer)
type URLAuditObserver struct {
	url    string
	client *resty.Client
}

// NewURLAuditObserver создаёт наблюдателя, отправляющего события на указанный url.
// Использует resty-клиент с таймаутом 5 секунд.
func NewURLAuditObserver(url string) *URLAuditObserver {
	client := resty.New().
		SetTimeout(5 * time.Second).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil {
				return true
			}
			code := r.StatusCode()
			return code == http.StatusTooManyRequests || code >= http.StatusInternalServerError
		}).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(8 * time.Second)

	return &URLAuditObserver{url: url, client: client}
}

// Notify отправляет событие POST-запросом на настроенный URL.
// Возвращает ошибку при сетевом сбое или если сервер вернул HTTP-статус ≥ 400.
func (o *URLAuditObserver) Notify(event AuditEvent) error {
	resp, err := o.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(event).
		Post(o.url)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("server returned status: %d", resp.StatusCode())
	}
	return nil
}
