package agent

import (
	"log"
	"time"

	models "github.com/F3dosik/metalert.git/internal/model"
)

func Run(endpoint string, reportInterval, pollInterval time.Duration) {

	metrics := &Metrics{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}

	// log.Printf("Агент запущен\nСервер: %s\nКонфигурация: {reportInterval: %v, pollInterval: %v}", endpoint, reportInterval, pollInterval)
	log.Printf("┌────────────────────────────────────────┐")
	log.Printf("│ Агент запущен                          │")
	log.Printf("│ Сервер: %-30s │", endpoint)
	log.Printf("│ Конфигурация:                          │")
	log.Printf("│    • PollInterval: %-19v │", pollInterval)
	log.Printf("│    • ReportInterval: %-17v │", reportInterval)
	log.Printf("└────────────────────────────────────────┘")

	sender := NewSender(endpoint)

	metrics.Update()

	sender.SendMetrics(metrics)

	tickerPoll := time.NewTicker(pollInterval)
	tickerReport := time.NewTicker(reportInterval)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	for {
		select {
		case <-tickerPoll.C:
			metrics.Update()
		case <-tickerReport.C:
			sender.SendMetrics(metrics)
		}
	}
}
