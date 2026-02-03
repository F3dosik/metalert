package agent

import (
	"context"
	"log"
	"sync"
	"time"
)

func Run(ctx context.Context, endpoint string, reportInterval, pollInterval time.Duration, rateLimit int) {

	// log.Printf("Агент запущен\nСервер: %s\nКонфигурация: {reportInterval: %v, pollInterval: %v}", endpoint, reportInterval, pollInterval)
	log.Printf("┌────────────────────────────────────────┐")
	log.Printf("│ Агент запущен                          │")
	log.Printf("│ Сервер: %-30s │", endpoint)
	log.Printf("│ Конфигурация:                          │")
	log.Printf("│    • PollInterval: %-19v │", pollInterval)
	log.Printf("│    • ReportInterval: %-17v │", reportInterval)
	log.Printf("└────────────────────────────────────────┘")

	metrics := NewMetrics()
	sender := NewSender(endpoint)

	sendCh := make(chan MetricsSnapshot, rateLimit*2)

	var wg sync.WaitGroup

	for i := 0; i < rateLimit; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for snapshot := range sendCh {
				if err := sender.SendMetrics(snapshot, "JSON", true); err != nil {
					log.Printf("send metrics error: %v", err)
				}
			}
		}(i)
	}

	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metrics.UpdateRuntimeMetrics()
				metrics.UpdateGopsutilMetrics()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(reportInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				snapshot := metrics.Snapshot()
				// Вложенный select нужен, чтобы горутина могла выйти при shutdown, даже если канал задач заполнен.
				select {
				case sendCh <- snapshot:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	<-ctx.Done()

	close(sendCh)
	wg.Wait()
}
