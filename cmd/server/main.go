package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/an3wers/notification-serv/internal/application/usecase"
	"github.com/an3wers/notification-serv/internal/infrastructure/email"
	"github.com/an3wers/notification-serv/internal/infrastructure/persistence/database"
	"github.com/an3wers/notification-serv/internal/pkg/config"
	"github.com/an3wers/notification-serv/internal/pkg/logger"
	"github.com/an3wers/notification-serv/internal/presentation/http/handlers"
	"github.com/an3wers/notification-serv/internal/presentation/http/router"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file")
	}

	// load config
	cfg := config.MustLoad()

	logg, err := logger.New(cfg.Logger.Level, cfg.Logger.Format)

	if err != nil {
		log.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	defer logg.Sync()

	logg.Info(
		"Init logger",
		zap.String("level", cfg.Logger.Level), zap.String("format", cfg.Logger.Format))

	// Connect to database
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		logg.Fatal("Failed to connect to database", zap.String("error", err.Error()))
	}
	defer db.Close()
	logg.Info("Connected to database")

	// repositories
	emailRepo := database.NewEmailRepository(db)

	// providers
	emailProvider := email.NewSMTPProvider(cfg.SMTP)

	// usecases
	sendEmailUC := usecase.NewSendEmailUseCase(emailRepo, emailProvider, cfg.SMTP, logg)
	getEmailStatusUC := usecase.NewGetEmailStatusUseCase(emailRepo)

	// init handlers
	healthHandler := handlers.NewHealthHandler(db.Pool)
	emailHandler := handlers.NewEmailHandler(sendEmailUC, getEmailStatusUC, cfg.Storage, logg)

	// setup chi router
	r := router.NewRouter(healthHandler, emailHandler, logg)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		logg.Info("Server started", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Fatal("Server failed", zap.String("error", err.Error()))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logg.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logg.Fatal("Server forced to shutdown", zap.String("error", err.Error()))
	}

	logg.Info("Server stopped")

}
