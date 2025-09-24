package agent

import (
	"time"

	models "github.com/F3dosik/metalert.git/internal/model"
)



func Run(endpoint string, reportInterval, pollInterval time.Duration) {
	
	metrics := &Metrics{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}

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
