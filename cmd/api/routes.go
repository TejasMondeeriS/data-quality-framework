package main

import (
	"net/http"

	chiprometheus "github.com/edjumacator/chi-prometheus"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(chiprometheus.NewMiddleware("Observability"))

	mux.NotFound(app.notFound)
	mux.MethodNotAllowed(app.methodNotAllowed)

	mux.Use(app.logAccess)
	mux.Use(app.recoverPanic)

	mux.Handle("/metrics", promhttp.Handler())

	// Swagger
	mux.Get("/swagger/*", httpSwagger.WrapHandler)

	// Health
	mux.Get("/health", app.HealthHandler)

	mux.Post("/query", app.AddQuery)
	mux.Get("/query", app.FetchAllQueries)

	mux.Post("/run", app.RunQuey)

	return mux
}
