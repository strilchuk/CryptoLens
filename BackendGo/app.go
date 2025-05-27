package main

import (
	"CryptoLens_Backend/container"
	"CryptoLens_Backend/env"
	"CryptoLens_Backend/initialization"
	"log"
	"net/http"
)

func main() {
	initialization.Initialize()
	
	ctr := container.NewContainer(initialization.DB, []byte(env.GetJWTSecret()))
	ctr.RegisterRoutes()

	log.Println("Server starting on " + env.GetServerPort())
	if err := http.ListenAndServe(":"+env.GetServerPort(), nil); err != nil {
		log.Fatal(err)
	}
}
