package models

import "time"

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
	TemplateCode string                 `json:"template_code"`
	Variables    UserData               `json:"variables"`
	RequestID    string                 `json:"request_id"`
	Priority     int                    `json:"priority"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// user-specific data for notification variables
type UserData struct {
	Name string                 `json:"name"`
	Link string                 `json:"link"`
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
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	DeviceTokens  []string               `json:"device_tokens"`
	Platform      string                 `json:"platform"` // "ios", "android", "web"
	TemplateID    string                 `json:"template_id,omitempty"`
	TemplateCode  string                 `json:"template_code,omitempty"`
	Title         string                 `json:"title"`
	Body          string                 `json:"body"`
	ImageURL      string                 `json:"image_url,omitempty"`
	Link          string                 `json:"link,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
	Variables     map[string]string      `json:"variables,omitempty"`
	Language      string                 `json:"language,omitempty"`
	Priority      string                 `json:"priority,omitempty"` // "high", "normal"
	AttemptCount  int                    `json:"attempt_count"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	RequestID     string                 `json:"request_id,omitempty"`
	ScheduledAt   *time.Time             `json:"scheduled_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
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
	AttemptCount    int                 `json:"attempt_count"`
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
	if n.Title == "" && n.Body == "" && n.TemplateID == "" && n.TemplateCode == "" {
		return ErrEmptyNotificationContent
	}
	return nil
}

// increments the attempt count
func (n *NotificationMessage) IncrementAttempt() {
	n.AttemptCount++
}

// determine if the notification should be retried
func (n *NotificationMessage) ShouldRetry(maxAttempts int) bool {
	return n.AttemptCount < maxAttempts
}
