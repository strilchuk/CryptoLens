package initialization

import (
	"SmallBot/env"
	"SmallBot/logger"
	"log"
)

func Initialize() {
	initLogger()
	env.Init()
}

func initLogger() {
	err := logger.Init("logs/app.log")
	if err != nil {
		log.Fatal(err)
	}
}
