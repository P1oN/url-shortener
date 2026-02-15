package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupAPIRoutes(handlers *Handlers) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/v1/health", handlers.HealthHandler).Methods(http.MethodGet)
	RegisterHandlers(router, handlers)
	router.HandleFunc("/v1/{code}", handlers.GetFullURLHandler).Methods(http.MethodGet)

	return router
}

func SetupRoutes(handlers *Handlers, enableSwagger bool) *mux.Router {
	router := setupAPIRoutes(handlers)
	if enableSwagger {
		router.HandleFunc("/swagger", handlers.SwaggerUIHandler).Methods(http.MethodGet)
		router.HandleFunc("/swagger/", handlers.SwaggerUIHandler).Methods(http.MethodGet)
		router.HandleFunc("/swagger/openapi.yaml", handlers.SwaggerSpecHandler).Methods(http.MethodGet)
	}
	// Public short links are generated as /{code}.
	router.HandleFunc("/{code}", handlers.GetFullURLHandler).Methods(http.MethodGet)

	return router
}
