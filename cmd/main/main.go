package main

import (
	"my-bank-service/internal/app"
	"my-bank-service/internal/config"
)

// @title Bank API
// @version 1.0
// @description API Server for

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	app.Run(config.ServerAddr, config.ServerPort)
}
