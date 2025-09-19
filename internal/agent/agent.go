package agent

import (
	"time"

	models "github.com/F3dosik/metalert.git/internal/model"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func Run(serverURL string) {
	metrics := &Metrics{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}

	sender := NewSender(serverURL)

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
