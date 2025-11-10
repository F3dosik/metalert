package agent

import (
	"log"
	"time"

	cfg "github.com/F3dosik/metalert.git/internal/config/agent"
	"github.com/F3dosik/metalert.git/pkg/models"
)

func Run(cfg *cfg.AgentConfig) {
	metrics := &Metrics{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}

	log.Printf("┌────────────────────────────────────────┐")
	log.Printf("│ Агент запущен                          │")
	log.Printf("│ Сервер: %-30s │", cfg.Endpoint)
	log.Printf("│ Конфигурация:                          │")
	log.Printf("│    • PollInterval: %-19v │", cfg.PollInterval)
	log.Printf("│    • ReportInterval: %-17v │", cfg.ReportInterval)
	log.Printf("└────────────────────────────────────────┘")

	sender := NewSender(cfg.Endpoint, cfg.Key)

	metrics.Update()

	sender.SendMetrics(metrics, "JSON", true)

	tickerPoll := time.NewTicker(cfg.PollInterval)
	tickerReport := time.NewTicker(cfg.ReportInterval)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	for {
		select {
		case <-tickerPoll.C:
			metrics.Update()
		case <-tickerReport.C:
			sender.SendMetrics(metrics, "JSON", true)
		}
	}
}
