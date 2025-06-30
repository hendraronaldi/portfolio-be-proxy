package main

import (
	"log"
	"net/http"

	"portfolio-be-proxy/config"
	"portfolio-be-proxy/routes"
)

const (
	defaultPort = "8080"
)

func main() {
	config.LoadConfig()

	mx := routes.SetupRoutes()

	srv := routes.ConfigureCors(mx, config.Config.FrontendOrigin)

	log.Printf("Server listening on port %s", config.Config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Config.Port, srv))
}
