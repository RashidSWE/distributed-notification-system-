package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	handler "github.com/zjoart/distributed-notification-system/push-service/internal/handlers"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/middleware"
)

type Server struct {
	router *mux.Router
	server *http.Server
}

func NewServer(
	host string,
	port int,
	healthHandler *handler.HealthHandler,
	notificationHandler *handler.NotificationHandler,
) *Server {
	router := mux.NewRouter()

	// middlewares

	router.Use(middleware.LoggingMiddleware)

	allowedOrigins := []string{
		"*",
	}
	router.Use(middleware.CorsMiddleware(allowedOrigins))

	// health check
	router.HandleFunc("/health", healthHandler.HandleHealth).Methods("GET")

	// Notification endpoints
	notifications := router.PathPrefix("/notifications").Subrouter()
	notifications.HandleFunc("/", notificationHandler.CreateNotification).Methods("POST")
	notifications.HandleFunc("/validate-tokens", notificationHandler.ValidateDeviceTokens).Methods("POST")

	// swagger documentation
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		router: router,
		server: server,
	}
}

func (s *Server) Start() error {
	logger.Info("Starting HTTP server", logger.Fields{
		"address": s.server.Addr,
	})

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}
