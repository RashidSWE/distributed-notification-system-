package handler

import (
	"context"
	"net/http"

	"github.com/zjoart/distributed-notification-system/push-service/internal/queue"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/handler"
)

type HealthHandler struct {
	queue *queue.RabbitMQ
	cache interface{ Health(context.Context) error }
}

func NewHealthHandler(queue *queue.RabbitMQ, cache interface{ Health(context.Context) error }) *HealthHandler {
	return &HealthHandler{
		queue: queue,
		cache: cache,
	}
}

func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health := map[string]string{
		"service": "push-service",
		"status":  "healthy",
	}

	// check rabbitmq
	if err := h.queue.Health(); err != nil {
		health["rabbitmq"] = "unhealthy: " + err.Error()
		health["status"] = "degraded"
	} else {
		health["rabbitmq"] = "connected"
	}

	// check redis
	if err := h.cache.Health(ctx); err != nil {
		health["redis"] = "unhealthy: " + err.Error()
		health["status"] = "degraded"
	} else {
		health["redis"] = "connected"
	}

	statusCode := http.StatusOK
	if health["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	handler.WriteJSON(w, statusCode, health)

}
