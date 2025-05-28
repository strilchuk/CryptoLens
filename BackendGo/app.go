package main

import (
	"CryptoLens_Backend/container"
	"CryptoLens_Backend/env"
	"CryptoLens_Backend/initialization"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	initialization.Initialize()
	
	ctr := container.NewContainer(initialization.DB, []byte(env.GetJWTSecret()))
	ctr.RegisterRoutes()

	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем фоновые задачи
	ctr.StartBackgroundTasks(ctx)

	// Настраиваем graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Println("Server starting on " + env.GetServerPort())
		if err := http.ListenAndServe(":"+env.GetServerPort(), nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Ждем сигнала для завершения
	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel() // Отменяем контекст, что приведет к остановке фоновых задач
}
