package audit

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type URLAuditObserver struct {
	url    string
	client *resty.Client
}

func NewURLAuditObserver(url string) *URLAuditObserver {
	return &URLAuditObserver{
		url:    url,
		client: resty.New().SetTimeout(5 * time.Second),
	}
}

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
