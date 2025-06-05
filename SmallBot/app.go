package main

import (
	"SmallBot/container"
	"SmallBot/env"
	"SmallBot/handlers"
	"SmallBot/initialization"
	"SmallBot/logger"
	"SmallBot/metrics"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	initialization.Initialize()

	ctr := container.NewContainer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctr.StartBackgroundTasks(ctx)

	metricsHandler := handlers.NewMetricsHandler()
	http.Handle("/metrics", metricsHandler)
	http.Handle("/metrics/summary", metricsHandler)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Server starting on " + env.GetServerPort())
		log.Println("📊 Метрики доступны на http://localhost:" + env.GetServerPort() + "/metrics")
		log.Println("📊 Сводка метрик: http://localhost:" + env.GetServerPort() + "/metrics/summary")

		log.Println("Server starting on " + env.GetServerPort())
		if err := http.ListenAndServe(":"+env.GetServerPort(), nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Периодический вывод метрик в лог
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logger.LogInfo(metrics.GetInstance().GetSummary())
			}
		}
	}()

	<-sigChan
	log.Println("Shutting down gracefully...")

	// Выводим финальные метрики
	logger.LogInfo("ФИНАЛЬНЫЕ МЕТРИКИ:")
	logger.LogInfo(metrics.GetInstance().GetSummary())

	cancel()
	if err := ctr.Close(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
