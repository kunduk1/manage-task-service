// Package api собирает HTTP-роутер сервиса.
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/kunduk1/manage-task-service/docs"
	authapi "github.com/kunduk1/manage-task-service/internal/api/auth"
	taskapi "github.com/kunduk1/manage-task-service/internal/api/task"
	teamapi "github.com/kunduk1/manage-task-service/internal/api/team"
	"github.com/kunduk1/manage-task-service/internal/token"
	"github.com/kunduk1/manage-task-service/internal/transport/middleware"
)

func NewRouter(authHandler *authapi.Handler, teamHandler *teamapi.Handler, taskHandler *taskapi.Handler, jwtManager *token.Manager) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(middleware.RequestLogger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Swagger UI at /swagger/index.html; spec served from the embedded docs package.
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(docs.SwaggerJSON)
	})
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	r.Route("/api/v1", func(r chi.Router) {
		// Публичные ручки аутентификации.
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		// Защищённые ручки: требуют валидный access-токен
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtManager))

			r.Post("/teams", teamHandler.Create)
			r.Get("/teams", teamHandler.List)
			r.Post("/teams/{id}/invite", teamHandler.Invite)

			r.Post("/tasks", taskHandler.Create)
			r.Get("/tasks", taskHandler.List)
			r.Put("/tasks/{id}", taskHandler.Update)
			r.Get("/tasks/{id}/history", taskHandler.History)
		})
	})

	return r
}
