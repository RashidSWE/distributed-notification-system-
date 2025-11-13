package models

import (
	"fmt"
	"time"
)

// status of a notification
type NotificationStatusEnum string

const (
	NotificationStatusDelivered NotificationStatusEnum = "delivered"
	NotificationStatusPending   NotificationStatusEnum = "pending"
	NotificationStatusFailed    NotificationStatusEnum = "failed"
)

// create a push notification request
type CreateNotificationRequest struct {
	UserID       string                 `json:"user_id"`
	DeviceTokens []string               `json:"device_tokens"`
	Platform     string                 `json:"platform,omitempty"` // "ios", "android", "web"
	Title        string                 `json:"title,omitempty"`
	Body         string                 `json:"body,omitempty"`
	ImageURL     string                 `json:"image_url,omitempty"`
	Link         string                 `json:"link,omitempty"`
	TemplateCode string                 `json:"template_code,omitempty"`
	Variables    UserData               `json:"variables,omitempty"`
	RequestID    string                 `json:"request_id"`
	Priority     int                    `json:"priority"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// user-specific data for notification variables
type UserData struct {
	Name string                 `json:"name"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// update notification status request
type UpdateNotificationStatusRequest struct {
	NotificationID string                 `json:"notification_id"`
	Status         NotificationStatusEnum `json:"status"`
	Timestamp      *time.Time             `json:"timestamp,omitempty"`
	Error          string                 `json:"error,omitempty"`
}

// notification message from queue
type NotificationMessage struct {
	ID               string            `json:"id"`
	NotificationType string            `json:"notification_type"` // "email", "push", "sms"
	UserID           string            `json:"user_id"`
	TemplateCode     string            `json:"template_code"`
	DeviceTokens     []string          `json:"device_tokens"`
	Variables        map[string]string `json:"variables,omitempty"`
	Platform         string            `json:"platform,omitempty"` // "ios", "android", "web"
	Priority         string            `json:"priority,omitempty"` // "high", "normal"
	CorrelationID    string            `json:"correlation_id,omitempty"`
	RequestID        string            `json:"request_id,omitempty"`
	ScheduledAt      *time.Time        `json:"scheduled_at,omitempty"`
	CreatedAt        time.Time         `json:"created_at,omitempty"`
}

type PushNotification struct {
	Title    string                 `json:"title"`
	Body     string                 `json:"body"`
	ImageURL string                 `json:"image_url,omitempty"`
	Link     string                 `json:"link,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Priority string                 `json:"priority,omitempty"`
}

type NotificationResult struct {
	MessageID     string    `json:"message_id"`
	DeviceToken   string    `json:"device_token"`
	Success       bool      `json:"success"`
	Error         string    `json:"error,omitempty"`
	SentAt        time.Time `json:"sent_at"`
	CorrelationID string    `json:"correlation_id,omitempty"`
}

// failed notification for the dead-letter queue
type FailedMessage struct {
	OriginalMessage NotificationMessage `json:"original_message"`
	Reason          string              `json:"reason"`
	FailedAt        time.Time           `json:"failed_at"`
	LastError       string              `json:"last_error"`
}

// device token validation result
type DeviceTokenValidation struct {
	Token  string `json:"token"`
	Valid  bool   `json:"valid"`
	Reason string `json:"reason,omitempty"`
}

type NotificationStatusResponse struct {
	ID            string                 `json:"id"`
	Status        NotificationStatusEnum `json:"status"`
	SentCount     int                    `json:"sent_count"`
	FailedCount   int                    `json:"failed_count"`
	LastUpdated   time.Time              `json:"last_updated"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	RequestID     string                 `json:"request_id,omitempty"`
}

// status queue message
type NotificationStatusMessage struct {
	NotificationID   string                 `json:"notification_id"`
	Status           NotificationStatusEnum `json:"status"`
	Timestamp        time.Time              `json:"timestamp"`
	Error            *string                `json:"error"`
	UserID           string                 `json:"user_id"`
	NotificationType string                 `json:"notification_type"`
	TemplateCode     string                 `json:"template_code"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// validates notification message
func (n *NotificationMessage) Validate() error {
	if n.ID == "" {
		return ErrInvalidNotificationID
	}
	if n.UserID == "" {
		return ErrInvalidUserID
	}
	if len(n.DeviceTokens) == 0 {
		return ErrNoDeviceTokens
	}
	if n.TemplateCode == "" {
		return fmt.Errorf("template_code is required")
	}
	if n.NotificationType == "" {
		return fmt.Errorf("notification_type is required")
	}
	if n.NotificationType != "push" {
		return fmt.Errorf("notification_type must be 'push', got '%s'", n.NotificationType)
	}
	return nil
}
