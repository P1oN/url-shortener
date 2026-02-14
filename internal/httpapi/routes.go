package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRoutes(handlers *Handlers) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/v1/health", handlers.HealthHandler).Methods(http.MethodGet)
	router.HandleFunc("/v1/shorten", handlers.CreateShortURLHandler).Methods(http.MethodPost)
	router.HandleFunc("/v1/{code}", handlers.GetFullURLHandler).Methods(http.MethodGet)

	return router
}
