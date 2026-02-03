package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/F3dosik/metalert.git/pkg/compression"
	"github.com/F3dosik/metalert.git/pkg/models"
	"github.com/go-resty/resty/v2"
)

type Sender struct {
	ServerURL string
	Client    *resty.Client
}

func NewSender(serverURL string) *Sender {
	client := resty.New()
	client.SetTimeout(5 * time.Second)
	return &Sender{
		ServerURL: serverURL,
		Client:    client,
	}
}

func (s *Sender) SendMetrics(snapshot MetricsSnapshot, sendType string, compress bool) error {
	switch sendType {
	case "URL":
		return s.sendMetricsIndividually(snapshot)
	case "JSON":
		return s.sendMetricsBatch(snapshot, compress)
	default:
		return fmt.Errorf("неизвестный тип отправки: %s", sendType)
	}
}

func (s *Sender) sendMetricURL(metricType models.MetricType, metricName, metricValue string) error {
	fullURL := s.prepareURL("/update/{metType}/{metName}/{metValue}")

	resp, err := s.Client.R().
		SetPathParams(map[string]string{
			"metType":  string(metricType),
			"metName":  metricName,
			"metValue": metricValue,
		}).
		SetHeader("Content-Type", "text/plain").
		Post(fullURL)

	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("неожиданный статус код: %d", resp.StatusCode())
	}

	log.Printf("Метрика %s успешно отправлена. Ответ: %s", metricName, resp.Body())
	return nil
}

func (s *Sender) sendMetricsIndividually(snapshot MetricsSnapshot) error {
	var errs []error

	metricType := models.TypeGauge
	for metricName, val := range snapshot.Gauges {
		metricValue := strconv.FormatFloat(float64(val), 'f', -1, 64)

		if err := s.sendMetricURL(metricType, metricName, metricValue); err != nil {
			errs = append(errs, fmt.Errorf("gauge %s %w", metricName, err))
		}
	}

	metricType = models.TypeCounter
	for metricName, val := range snapshot.Counters {
		metricValue := strconv.Itoa(int(val))

		if err := s.sendMetricURL(metricType, metricName, metricValue); err != nil {
			errs = append(errs, fmt.Errorf("counter %s %w", metricName, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("send metrics individually failed: %v", errs)
	}

	return nil
}

func (s *Sender) sendMetricsBatch(snapshot MetricsSnapshot, compress bool) error {
	var metrics []models.Metric
	for id, v := range snapshot.Gauges {
		value := v
		metric := models.Metric{
			ID:    id,
			MType: models.TypeGauge,
			Value: &value,
		}
		metrics = append(metrics, metric)
	}
	for id, d := range snapshot.Counters {
		delta := d
		metric := models.Metric{
			ID:    id,
			MType: models.TypeCounter,
			Delta: &delta,
		}
		metrics = append(metrics, metric)
	}
	return s.sendMetricJSON(metrics, compress)

}
func (s *Sender) sendMetricJSON(metrics []models.Metric, compress bool) error {
	fullURL := s.prepareURL("/updates/")

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	if compress {
		if jsonData, err = compression.Compress(jsonData); err != nil {
			return fmt.Errorf("ошибка сжатия: %w", err)
		}
	}

	req := s.Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonData)

	if compress {
		req.SetHeader("Content-Encoding", "gzip")
	}

	resp, err := req.Post(fullURL)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("ошибка сервера: %s", resp.Status())
	}

	log.Printf("Отправлено %d метрик (%s). Ответ: %s",
		len(metrics),
		map[bool]string{true: "gzip", false: "plain"}[compress],
		resp.Body())

	return nil
}

func (s *Sender) prepareURL(path string) string {
	fullURL, err := url.JoinPath(s.ServerURL, path)
	if err != nil {
		fullURL = s.ServerURL + path
	}

	parsed, err := url.Parse(fullURL)
	if err != nil {
		return "http://" + fullURL
	}

	if parsed.Scheme == "" {
		parsed.Scheme = "http"
		fullURL = parsed.String()
	}
	return fullURL
}
