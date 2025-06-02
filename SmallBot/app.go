package main

import (
	"SmallBot/container"
	"SmallBot/env"
	"SmallBot/initialization"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	initialization.Initialize()

	ctr := container.NewContainer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctr.StartBackgroundTasks(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Server starting on " + env.GetServerPort())
		if err := http.ListenAndServe(":"+env.GetServerPort(), nil); err != nil {
			log.Fatal(err)
		}
	}()

	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel()
	if err := ctr.Close(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
