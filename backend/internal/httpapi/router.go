package httpapi

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/app"
)

func NewRouter(service *app.Service, allowedOrigins []string, logger *slog.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(recoverer(logger))
	router.Use(requestLogger(logger))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-API-Version"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		writeErrorResponse(w, http.StatusNotFound, "not_found", "Route not found")
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
	})
	router.Route("/api", func(api chi.Router) {
		api.Use(apiVersion)
		api.Get("/connections", listConnections(service))
		api.Post("/connections", createConnection(service))
		api.Post("/connections/test", testConnection(service))
		api.Delete("/connections/{connectionID}", deleteConnection(service))
		api.Get("/connections/{connectionID}/tables", listTables(service))
		api.Get("/connections/{connectionID}/tables/{tableName}", describeTable(service))
		api.Get("/connections/{connectionID}/tables/{tableName}/items", listItems(service))
		api.Put("/connections/{connectionID}/tables/{tableName}/items", putItem(service))
		api.Delete("/connections/{connectionID}/tables/{tableName}/items", deleteItem(service))
	})
	return router
}

func apiVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := strings.TrimSpace(r.Header.Get("X-API-Version"))
		if version != "v1" {
			writeErrorResponse(w, http.StatusBadRequest, "unsupported_api_version", "X-API-Version must be v1")
			return
		}
		next.ServeHTTP(w, r)
	})
}
