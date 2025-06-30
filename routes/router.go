package routes

import (
	"portfolio-be-proxy/handlers"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	// Apply middleware
	r := mux.NewRouter()

	// Apply middleware to all routes under /api
	api := r.PathPrefix("/api").Subrouter()
	api.Use(loggingMiddleware)
	api.Use(apiKeyMiddleware)
	api.Use(globalRateLimitMiddleware)
	api.HandleFunc("/agent/resume", handlers.QueryCVHandler)

	return r
}
