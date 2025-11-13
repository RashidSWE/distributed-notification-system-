package template

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type PushTemplate struct {
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Category    string                 `json:"category,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Body        string                 `json:"body,omitempty"`
	ImageURL    string                 `json:"image_url,omitempty"`
	IconURL     string                 `json:"icon_url,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Color       string                 `json:"color,omitempty"`
	Sound       string                 `json:"sound,omitempty"`
	Badge       int                    `json:"badge,omitempty"`
	Priority    int                    `json:"priority,omitempty"`
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) GetPushTemplate(ctx context.Context, templateCode string) (*PushTemplate, error) {
	url := fmt.Sprintf("%s/templates/push/%s", c.baseURL, templateCode)

	logFields := logger.Fields{
		"template_code": templateCode,
		"url":           url,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	logger.Info("Fetching push template from template service", logFields)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to fetch template from template service", logger.Merge(
			logger.WithError(err),
			logFields,
		))
		return nil, fmt.Errorf("template service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		logger.Warn("Template not found in template service", logger.Merge(
			logFields,
			logger.Fields{"status_code": resp.StatusCode},
		))
		return nil, fmt.Errorf("template not found: %s", templateCode)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Template service returned error", logger.Merge(
			logFields,
			logger.Fields{"status_code": resp.StatusCode},
		))
		return nil, fmt.Errorf("template service returned status %d", resp.StatusCode)
	}

	var template PushTemplate
	if err := json.NewDecoder(resp.Body).Decode(&template); err != nil {
		logger.Error("Failed to decode template response", logger.Merge(
			logger.WithError(err),
			logFields,
		))
		return nil, fmt.Errorf("failed to decode template: %w", err)
	}

	logger.Info("Successfully fetched template", logger.Merge(
		logFields,
		logger.Fields{"template_name": template.Name},
	))

	return &template, nil
}
