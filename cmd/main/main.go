package main

import (
	"my-bank-service/internal/app"
	"my-bank-service/internal/config"
)

func main() {
	app.Run(config.ServerAddr, config.ServerPort)
}
