package push

import (
	"context"
	"fmt"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/zjoart/distributed-notification-system/push-service/internal/models"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"
	"google.golang.org/api/option"
)

type FCMService struct {
	client         *messaging.Client
	timeout        time.Duration
	circuitBreaker *CircuitBreaker
}

func NewFCMService(ctx context.Context, projectID, credentialsPath string, timeout int, cb *CircuitBreaker) (*FCMService, error) {
	opt := option.WithCredentialsFile(credentialsPath)

	config := &firebase.Config{
		ProjectID: projectID,
	}

	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize FCM client: %w", err)
	}

	logger.Info("FCM service initialized successfully", logger.Fields{
		"project_id": projectID,
	})

	return &FCMService{
		client:         client,
		timeout:        time.Duration(timeout) * time.Second,
		circuitBreaker: cb,
	}, nil
}

// send a push notification to a single device
func (s *FCMService) SendNotification(ctx context.Context, deviceToken string, notification *models.PushNotification) (*models.NotificationResult, error) {

	// trim whitespace from token
	deviceToken = strings.TrimSpace(deviceToken)

	result := &models.NotificationResult{
		DeviceToken: deviceToken,
		SentAt:      time.Now(),
	}

	// check circuit breaker
	if err := s.circuitBreaker.Call(func() error {
		// create timeout context
		ctx, cancel := context.WithTimeout(ctx, s.timeout)
		defer cancel()

		// build FCM message
		message := &messaging.Message{
			Token: deviceToken,
			Notification: &messaging.Notification{
				Title:    notification.Title,
				Body:     notification.Body,
				ImageURL: notification.ImageURL,
			},
			Webpush: &messaging.WebpushConfig{
				Notification: &messaging.WebpushNotification{
					Title: notification.Title,
					Body:  notification.Body,
					Icon:  notification.ImageURL,
				},
				FCMOptions: &messaging.WebpushFCMOptions{
					Link: notification.Link,
				},
			},
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					Title:       notification.Title,
					Body:        notification.Body,
					ClickAction: notification.Link,
				},
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Alert: &messaging.ApsAlert{
							Title: notification.Title,
							Body:  notification.Body,
						},
						Sound: "default",
					},
				},
			},
			Data: convertDataToString(notification.Data),
		}

		// set priority
		if notification.Priority == "high" {
			message.Android.Priority = "high"
		}

		// send message
		messageID, err := s.client.Send(ctx, message)
		if err != nil {
			result.Success = false
			result.Error = err.Error()

			logger.Error("Failed to send FCM notification", logger.Fields{
				"device_token": deviceToken,
				"error":        err.Error(),
			})

			return err
		}

		result.Success = true
		result.MessageID = messageID

		logger.Info("FCM notification sent successfully", logger.Fields{
			"message_id":   messageID,
			"device_token": deviceToken,
		})

		return nil
	}); err != nil {
		if err == models.ErrCircuitBreakerOpen {
			result.Success = false
			result.Error = "FCM service temporarily unavailable"
			logger.Warn("Circuit breaker is open, FCM service unavailable")
		}
		return result, err
	}

	return result, nil
}

// sends notification to multiple devices
func (s *FCMService) SendToMultipleDevices(ctx context.Context, deviceTokens []string, notification *models.PushNotification) ([]*models.NotificationResult, error) {

	// trim whitespace from all tokens
	for i := range deviceTokens {
		deviceTokens[i] = strings.TrimSpace(deviceTokens[i])
	}

	results := make([]*models.NotificationResult, 0, len(deviceTokens))

	// check circuit breaker
	if err := s.circuitBreaker.Call(func() error {

		// create timeout context
		ctx, cancel := context.WithTimeout(ctx, s.timeout)
		defer cancel()

		// build multiple message
		message := &messaging.MulticastMessage{
			Tokens: deviceTokens,
			Notification: &messaging.Notification{
				Title:    notification.Title,
				Body:     notification.Body,
				ImageURL: notification.ImageURL,
			},
			Webpush: &messaging.WebpushConfig{
				Notification: &messaging.WebpushNotification{
					Title: notification.Title,
					Body:  notification.Body,
					Icon:  notification.ImageURL,
				},
				FCMOptions: &messaging.WebpushFCMOptions{
					Link: notification.Link,
				},
			},
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					Title:       notification.Title,
					Body:        notification.Body,
					ClickAction: notification.Link,
				},
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Alert: &messaging.ApsAlert{
							Title: notification.Title,
							Body:  notification.Body,
						},
						Sound: "default",
					},
				},
			},
			Data: convertDataToString(notification.Data),
		}

		batchResponse, err := s.client.SendEachForMulticast(ctx, message)
		if err != nil {
			logger.Error("Failed to send FCM multicast", logger.WithError(err))
			return err
		}

		if batchResponse.Responses == nil || len(batchResponse.Responses) != len(deviceTokens) {
			logger.Error("FCM multicast response mismatch", logger.Fields{
				"responses_nil": batchResponse.Responses == nil,
				"responses_len": len(batchResponse.Responses),
				"tokens_len":    len(deviceTokens),
			})

			// mark all as failed
			for _, token := range deviceTokens {
				results = append(results, &models.NotificationResult{
					DeviceToken: token,
					Success:     false,
					Error:       "FCM multicast response mismatch",
					SentAt:      time.Now(),
				})
			}
		} else {

			for i, resp := range batchResponse.Responses {
				result := &models.NotificationResult{
					DeviceToken: deviceTokens[i],
					SentAt:      time.Now(),
				}

				if resp.Success {
					result.Success = true
					result.MessageID = resp.MessageID
				} else {
					result.Success = false
					result.Error = resp.Error.Error()

					logger.Error("Failed to send to device", logger.Fields{
						"device_token": deviceTokens[i],
						"error":        resp.Error.Error(),
					})
				}

				results = append(results, result)
			}
		}
		logger.Info("FCM multicast sent", logger.Fields{
			"success_count": batchResponse.SuccessCount,
			"failure_count": batchResponse.FailureCount,
		})

		return nil
	}); err != nil {
		if err == models.ErrCircuitBreakerOpen {

			// create failed results for all tokens
			for _, token := range deviceTokens {
				results = append(results, &models.NotificationResult{
					DeviceToken: token,
					Success:     false,
					Error:       "FCM service temporarily unavailable",
					SentAt:      time.Now(),
				})
			}
			logger.Warn("Circuit breaker is open, FCM service unavailable")
		}
		return results, err
	}

	return results, nil
}

// validates if a device token is valid
func (s *FCMService) ValidateDeviceToken(ctx context.Context, deviceToken string) (*models.DeviceTokenValidation, error) {
	validation := &models.DeviceTokenValidation{
		Token: deviceToken,
	}

	// FCM doesn't have a direct validation endpoint
	// so I'm validating by attempting a dry run send
	message := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: "Validation",
			Body:  "Token validation",
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// dry run to validate token
	_, err := s.client.SendDryRun(ctx, message)
	if err != nil {
		validation.Valid = false
		validation.Reason = err.Error()
		return validation, nil
	}

	validation.Valid = true
	return validation, nil
}

// converts map[string]interface{} to map[string]string for FCM
func convertDataToString(data map[string]interface{}) map[string]string {
	if data == nil {
		return nil
	}

	result := make(map[string]string)
	for key, value := range data {
		result[key] = fmt.Sprintf("%v", value)
	}
	return result
}
