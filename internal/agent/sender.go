package agent

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/F3dosik/metalert/internal/crypto"
	pb "github.com/F3dosik/metalert/internal/proto"
	"github.com/F3dosik/metalert/pkg/compression"
	"github.com/F3dosik/metalert/pkg/models"
)

// Sender отвечает за отправку метрик на сервер.
//
// Поддерживает три режима через [Sender.SendMetrics]:
//   - "URL"  — каждая метрика отдельным POST /update/{type}/{name}/{value}
//   - "JSON" — все метрики одним POST /updates/ (опционально gzip)
//   - "GRPC" — батч метрик через gRPC UpdateMetrics
type Sender struct {
	// ServerURL — базовый адрес сервера метрик, например "http://localhost:8080".
	ServerURL string

	// Client — HTTP-клиент resty с настроенным таймаутом.
	Client *resty.Client

	// CryptoKey — публичный ключ для поддержки асимметричного шифрования.
	CryptoKey *rsa.PublicKey

	// LocalIP — IP-адрес хоста агента, передаётся в заголовке X-Real-IP / gRPC metadata.
	LocalIP string

	// grpcClient — клиент gRPC MetricsService (nil, если gRPC не настроен).
	grpcClient pb.MetricsClient
}

// NewSender создаёт Sender, настроенный на отправку метрик.
// Таймаут HTTP-запросов — 5 секунд.
// Если grpcEndpoint не пустой, создаётся ленивое gRPC-соединение.
func NewSender(serverURL, grpcEndpoint string, publicKey *rsa.PublicKey) *Sender {
	client := resty.New()
	client.SetTimeout(5 * time.Second)

	s := &Sender{
		ServerURL: serverURL,
		Client:    client,
		CryptoKey: publicKey,
		LocalIP:   resolveLocalIP(serverURL),
	}

	if grpcEndpoint != "" {
		conn, err := grpc.NewClient(grpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Warning: failed to create gRPC client: %v", err)
		} else {
			s.grpcClient = pb.NewMetricsClient(conn)
		}
	}

	return s
}

// resolveLocalIP определяет локальный IP-адрес, используемый для соединения с сервером.
// Использует UDP-соединение (без отправки пакетов) для выбора нужного интерфейса.
func resolveLocalIP(serverURL string) string {
	parsed, err := url.Parse(serverURL)
	if err != nil {
		return ""
	}
	host := parsed.Host
	if host == "" {
		return ""
	}
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = host + ":80"
	}
	conn, err := net.Dial("udp", host)
	if err != nil {
		return ""
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

// SendMetrics отправляет все метрики из memMetrics на сервер.
//
// Параметры:
//   - sendType: "URL" — поштучная отправка, "JSON" — пакетная
//   - compress: при sendType="JSON" сжимает тело запроса через gzip
//
// Захватывает RLock на время отправки, не блокируя параллельный сбор метрик.
func (s *Sender) SendMetrics(memMetrics *Metrics, sendType string, compress bool) {
	memMetrics.mu.RLock()
	defer memMetrics.mu.RUnlock()

	switch sendType {
	case "URL":
		s.sendMetricsIndividually(memMetrics)
	case "JSON":
		s.sendMetricsBatch(memMetrics, compress)
	case "GRPC":
		s.sendMetricsGRPC(memMetrics)
	default:
		log.Printf("Неизвестный тип отправки: %s", sendType)
	}
}

// sendMetricURL отправляет одну метрику через URL-параметры.
// Маршрут: POST /update/{metType}/{metName}/{metValue}
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

// sendMetricsIndividually отправляет каждую метрику отдельным HTTP-запросом.
// Используется при sendType="URL".
func (s *Sender) sendMetricsIndividually(memMetrics *Metrics) {
	for metricName, val := range memMetrics.Gauges {
		metricValue := strconv.FormatFloat(float64(val), 'f', -1, 64)
		if err := s.sendMetricURL(models.TypeGauge, metricName, metricValue); err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", metricName, err)
		}
	}
	for metricName, val := range memMetrics.Counters {
		metricValue := strconv.Itoa(int(val))
		if err := s.sendMetricURL(models.TypeCounter, metricName, metricValue); err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", metricName, err)
		}
	}
}

// sendMetricsBatch собирает все метрики в срез и отправляет одним JSON-запросом.
// Используется при sendType="JSON".
func (s *Sender) sendMetricsBatch(memMetrics *Metrics, compress bool) {
	var metrics []models.Metric
	for id, v := range memMetrics.Gauges {
		value := v
		metrics = append(metrics, models.Metric{
			ID:    id,
			MType: models.TypeGauge,
			Value: &value,
		})
	}
	for id, d := range memMetrics.Counters {
		delta := d
		metrics = append(metrics, models.Metric{
			ID:    id,
			MType: models.TypeCounter,
			Delta: &delta,
		})
	}
	if err := s.sendMetricJSON(metrics, compress); err != nil {
		log.Printf("Ошибка отправки метрик: %v", err)
	}
}

// sendMetricJSON отправляет срез метрик одним POST-запросом на /updates/.
//
// При compress=true сжимает JSON-тело через gzip и устанавливает
// заголовок Content-Encoding: gzip.
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

	body, encrypted, err := s.encryptIfNeeded(jsonData)
	if err != nil {
		return fmt.Errorf("ошибка шифрования: %w", err)
	}

	req := s.Client.R().SetBody(body)
	if encrypted {
		req.SetHeader("X-Encrypted", "true")
		req.SetHeader("Content-Type", "application/octet-stream")
	} else {
		req.SetHeader("Content-Type", "application/json")
	}

	if compress {
		req.SetHeader("Content-Encoding", "gzip")
	}

	req.SetHeader("X-Real-IP", s.LocalIP)

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

// prepareURL формирует полный URL из базового адреса сервера и пути.
// Если схема не указана, автоматически добавляет "http://".
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

// sendMetricsGRPC отправляет все метрики батчем через gRPC UpdateMetrics.
// IP-адрес агента передаётся в метаданных запроса с ключом x-real-ip.
func (s *Sender) sendMetricsGRPC(memMetrics *Metrics) {
	if s.grpcClient == nil {
		log.Printf("gRPC клиент не инициализирован")
		return
	}

	pbMetrics := make([]*pb.Metric, 0, len(memMetrics.Gauges)+len(memMetrics.Counters))
	for id, v := range memMetrics.Gauges {
		id, v := id, v
		val := float64(v)
		mtype := pb.Metric_GAUGE
		pbMetrics = append(pbMetrics, pb.Metric_builder{
			Id:    &id,
			Type:  &mtype,
			Value: &val,
		}.Build())
	}
	for id, d := range memMetrics.Counters {
		id, d := id, d
		delta := int64(d)
		mtype := pb.Metric_COUNTER
		pbMetrics = append(pbMetrics, pb.Metric_builder{
			Id:    &id,
			Type:  &mtype,
			Delta: &delta,
		}.Build())
	}

	req := pb.UpdateMetricsRequest_builder{Metrics: pbMetrics}.Build()

	md := metadata.Pairs("x-real-ip", s.LocalIP)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	if _, err := s.grpcClient.UpdateMetrics(ctx, req); err != nil {
		log.Printf("gRPC UpdateMetrics error: %v", err)
		return
	}

	log.Printf("gRPC: отправлено %d метрик", len(pbMetrics))
}

func (s *Sender) encryptIfNeeded(data []byte) ([]byte, bool, error) {
	if s.CryptoKey == nil {
		return data, false, nil
	}
	encrypted, err := crypto.Encrypt(data, s.CryptoKey)
	if err != nil {
		return nil, false, err
	}
	return encrypted, true, nil
}
