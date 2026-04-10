// Package agent реализует агент сбора и отправки метрик на сервер.
//
// Агент периодически собирает runtime-метрики процесса ([Metrics.Update])
// и отправляет их на сервер ([Sender.SendMetrics]) по двум интервалам:
//   - pollInterval — частота сбора метрик
//   - reportInterval — частота отправки на сервер
//
// Поддерживаются два режима отправки:
//   - "URL" — каждая метрика отправляется отдельным POST-запросом
//   - "JSON" — все метрики отправляются одним пакетным запросом (опционально со сжатием gzip)
package agent

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/F3dosik/metalert/internal/crypto"
	"github.com/F3dosik/metalert/pkg/models"
)

// Run запускает агент сбора и отправки метрик.
//
// Сразу после старта выполняет первый сбор и отправку метрик,
// затем запускает два независимых тикера:
//   - каждые pollInterval собирает текущие метрики runtime
//   - каждые reportInterval отправляет накопленные метрики на сервер
//
// Блокирует выполнение горутины навсегда (рассчитан на запуск в main).
//
// Пример:
//
//	agent.Run("http://localhost:8080", 10*time.Second, 2*time.Second)
func Run(endpoint string, reportInterval, pollInterval time.Duration, cryptoKey, grpcEndpoint string) {
	metrics := &Metrics{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}

	log.Printf("┌────────────────────────────────────────┐")
	log.Printf("│ Агент запущен                          │")
	log.Printf("│ Сервер: %-30s │", endpoint)
	log.Printf("│ Конфигурация:                          │")
	log.Printf("│    • PollInterval: %-19v │", pollInterval)
	log.Printf("│    • ReportInterval: %-17v │", reportInterval)
	log.Printf("└────────────────────────────────────────┘")

	publicKey, err := crypto.LoadPublicKey(cryptoKey)
	if err != nil {
		log.Printf("%s", err.Error())
	}

	sender := NewSender(endpoint, grpcEndpoint, publicKey)

	sendType := "JSON"
	if grpcEndpoint != "" {
		sendType = "GRPC"
	}

	metrics.Update()
	sender.SendMetrics(metrics, sendType, true)

	tickerPoll := time.NewTicker(pollInterval)
	tickerReport := time.NewTicker(reportInterval)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	for {
		select {
		case <-tickerPoll.C:
			metrics.Update()
		case <-tickerReport.C:
			sender.SendMetrics(metrics, sendType, true)
		case <-sigs:
			log.Print("Получен сигнал завершения, отправляем имеющиеся метрики...")
			sender.SendMetrics(metrics, sendType, true)
			log.Print("Aгент завершен")
			return
		}
	}
}
