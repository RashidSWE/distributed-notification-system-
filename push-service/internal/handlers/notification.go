package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/zjoart/distributed-notification-system/push-service/internal/models"
	"github.com/zjoart/distributed-notification-system/push-service/internal/queue"
	"github.com/zjoart/distributed-notification-system/push-service/internal/service"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/handler"
)

type NotificationHandler struct {
	service   *service.NotificationService
	queue     *queue.RabbitMQ
	validator *validator.Validate
}

func NewNotificationHandler(service *service.NotificationService, queue *queue.RabbitMQ) *NotificationHandler {
	return &NotificationHandler{
		service:   service,
		queue:     queue,
		validator: validator.New(),
	}
}

func (h *NotificationHandler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	var req models.CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// validate request
	if err := h.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		handler.RespondWithValidationError(w, validationErrors)
		return
	}

	variables := map[string]string{
		"name": req.Variables.Name,
		"link": req.Variables.Link,
	}
	for key, val := range req.Variables.Meta {
		if strVal, ok := val.(string); ok {
			variables[key] = strVal
		}
	}

	message := &models.NotificationMessage{
		ID:           req.RequestID,
		UserID:       req.UserID,
		DeviceTokens: req.DeviceTokens,
		Platform:     req.Platform,
		Title:        req.Title,
		Body:         req.Body,
		ImageURL:     req.ImageURL,
		Link:         req.Link,
		TemplateCode: req.TemplateCode,
		Variables:    variables,
		Priority:     priorityToString(req.Priority),
		RequestID:    req.RequestID,
		Data:         req.Metadata,
		CreatedAt:    time.Now(),
	}

	// push message to queue
	if err := h.queue.Publish(r.Context(), "push.queue", message); err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to queue notification", err)
		return
	}

	response := map[string]interface{}{
		"notification_id": req.RequestID,
		"status":          "queued",
		"message":         "Notification queued successfully",
	}

	handler.RespondWithSuccessAndStatus(w, http.StatusAccepted, "Notification queued successfully", response)
}

func (h *NotificationHandler) ValidateDeviceTokens(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Tokens []string `json:"tokens"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(req.Tokens) == 0 {
		handler.RespondWithError(w, http.StatusBadRequest, "No tokens provided", nil)
		return
	}

	validations, err := h.service.ValidateDeviceTokens(r.Context(), req.Tokens)
	if err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to validate tokens", err)
		return
	}

	handler.RespondWithSuccess(w, "Tokens validated successfully", validations)
}

func priorityToString(priority int) string {
	if priority >= 5 {
		return "high"
	}
	return "normal"
}
