package agent

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	models "github.com/F3dosik/metalert.git/internal/model"
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

func (s *Sender) SendMetrics(metrics *Metrics) {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	// Отправка gauge метрик
	metricType := models.TypeGauge
	for metricName, val := range metrics.Gauges {
		metricValue := strconv.FormatFloat(float64(val), 'f', -1, 64)

		err := s.sendMetric(metricType, metricName, metricValue)
		if err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", metricName, err)
		}
	}

	// Отправка counter метрик
	metricType = models.TypeCounter
	for metricName, val := range metrics.Counters {
		metricValue := fmt.Sprint(val)

		err := s.sendMetric(metricType, metricName, metricValue)
		if err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", metricName, err)
		}
	}
}

func (s *Sender) sendMetric(metricType models.MetricType, metricName, metricValue string) error {
	fullURL := s.ServerURL + "/update/{metType}/{metName}/{metValue}"

	if !strings.HasPrefix(fullURL, "http://") && !strings.HasPrefix(fullURL, "https://") {
		fullURL = "http://" + fullURL
	}

	// log.Printf("Отправка метрики: %s %s=%s", metricType, metricName, metricValue)

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

	log.Printf("Метрика %s успешно отправлена. Ответ: %s", metricName, resp.Body()) // Или использовать resp.String и проверять ошибку
	return nil
}
