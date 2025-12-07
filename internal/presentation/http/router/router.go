package router

import (
	"time"

	"github.com/an3wers/notification-serv/internal/pkg/logger"
	"github.com/an3wers/notification-serv/internal/presentation/http/handlers"
	"github.com/an3wers/notification-serv/internal/presentation/http/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	healthHandler *handlers.HealthHandler,
	emailHandler *handlers.EmailHandler,
	log *logger.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS())
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", healthHandler.Health)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/emails", func(r chi.Router) {
			r.Post("/", emailHandler.SendEmail)
			r.Get("/{id}", emailHandler.GetEmailStatus)
		})
	})

	return r
}
