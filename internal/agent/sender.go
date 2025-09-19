package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	models "github.com/F3dosik/metalert.git/internal/model"
)

type Sender struct {
	ServerURL string
	Client    *http.Client
}

func NewSender(serverURL string) *Sender {
	return &Sender{
		ServerURL: serverURL,
		Client:    &http.Client{},
	}
}

func (s *Sender) SendMetrics(metrics *Metrics) {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	client := &http.Client{Timeout: 5 * time.Second}

	// Отправка gauge метрик
	metricType := models.GaugeType
	for metricName, val := range metrics.Gauges {
		metricValue := strconv.FormatFloat(float64(val), 'f', -1, 64)

		err := s.sendMetric(client, metricType, metricName, metricValue)
		if err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", metricName, err)
		}
	}

	// Отправка counter метрик
	metricType = models.CounterType
	for metricName, val := range metrics.Counters {
		metricValue := fmt.Sprint(val)

		err := s.sendMetric(client, metricType, metricName, metricValue)
		if err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", metricName, err)
		}
	}
}

func (s *Sender) sendMetric(client *http.Client, metricType, metricName, metricValue string) error {
	baseURL, _ := url.Parse(s.ServerURL)
	baseURL.Path = path.Join(baseURL.Path, "update", metricType, url.PathEscape(metricName), metricValue)
	fullURL := baseURL.String()

	req, err := http.NewRequest(http.MethodPost, fullURL, nil)
	fmt.Println(fullURL)
	if err != nil {
	
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	log.Printf("-> Отправляю %s %s=%s", metricType, metricName, metricValue)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неожиданный статус код: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	log.Printf("Метрика %s успешно отправлена. Ответ: %s", metricName, string(body))
	return nil
}
