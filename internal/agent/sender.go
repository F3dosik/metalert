package agent

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	Key       string
}

func NewSender(serverURL, key string) *Sender {
	client := resty.New()
	client.SetTimeout(5 * time.Second)
	return &Sender{
		ServerURL: serverURL,
		Client:    client,
		Key:       key,
	}
}

func (s *Sender) SendMetrics(memMetrics *Metrics, sendType string, compress bool) {
	memMetrics.mu.RLock()
	defer memMetrics.mu.RUnlock()

	switch sendType {
	case "URL":
		s.sendMetricsIndividually(memMetrics)
	case "JSON":
		s.sendMetricsBatch(memMetrics, compress)
	default:
		log.Printf("Неизвестный тип отправки: %s", sendType)
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

func (s *Sender) sendMetricsIndividually(memMetrics *Metrics) {
	metricType := models.TypeGauge
	for metricName, val := range memMetrics.Gauges {
		metricValue := strconv.FormatFloat(float64(val), 'f', -1, 64)

		err := s.sendMetricURL(metricType, metricName, metricValue)
		if err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", metricName, err)
		}
	}
	metricType = models.TypeCounter
	for metricName, val := range memMetrics.Counters {
		metricValue := strconv.Itoa(int(val))

		err := s.sendMetricURL(metricType, metricName, metricValue)
		if err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", metricName, err)
		}
	}
}

func (s *Sender) sendMetricsBatch(memMetrics *Metrics, compress bool) {
	var metrics []models.Metric
	for id, v := range memMetrics.Gauges {
		value := v
		metric := models.Metric{
			ID:    id,
			MType: models.TypeGauge,
			Value: &value,
		}
		metrics = append(metrics, metric)
	}
	for id, d := range memMetrics.Counters {
		delta := d
		metric := models.Metric{
			ID:    id,
			MType: models.TypeCounter,
			Delta: &delta,
		}
		metrics = append(metrics, metric)
	}
	err := s.sendMetricJSON(metrics, compress)
	if err != nil {
		log.Printf("Ошибка отправки метрик: %v", err)
	}
}
func (s *Sender) sendMetricJSON(metrics []models.Metric, compress bool) error {
	fullURL := s.prepareURL("/updates/")

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}
	var hash string
	if s.Key != "" {
		h := hmac.New(sha256.New, []byte(s.Key))
		h.Write(jsonData)
		hash = hex.EncodeToString(h.Sum(nil))
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

	if s.Key != "" {
		req.SetHeader("HashSHA256", hash)
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
