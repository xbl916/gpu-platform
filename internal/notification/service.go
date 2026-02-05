package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gpu-platform/internal/config"
)

type NotificationService struct {
	cfg *config.NotificationConfig
}

type Notification struct {
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Recipient string                 `json:"recipient"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	SentAt    time.Time              `json:"sentAt"`
}

type EmailNotification struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	HTML    bool   `json:"html"`
}

func NewNotificationService(cfg *config.NotificationConfig) *NotificationService {
	return &NotificationService{cfg: cfg}
}

func (s *NotificationService) SendEmail(ctx context.Context, email *EmailNotification) error {
	if !s.cfg.Enabled || s.cfg.Email.SMTPHost == "" {
		fmt.Printf("Email notification (mock): To=%s, Subject=%s\n", email.To, email.Subject)
		return nil
	}

	fmt.Printf("Sending email to %s: %s\n", email.To, email.Subject)
	return nil
}

func (s *NotificationService) SendWebhook(ctx context.Context, payload map[string]interface{}) error {
	if !s.cfg.Enabled || !s.cfg.Webhook.Enabled {
		fmt.Printf("Webhook notification (mock): URL=%s\n", s.cfg.Webhook.URL)
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.cfg.Webhook.URL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	return nil
}

func (s *NotificationService) NotifyContainerStarted(ctx context.Context, userEmail, containerID, containerName string) error {
	return s.SendEmail(ctx, &EmailNotification{
		To:      userEmail,
		Subject: "容器已启动 - " + containerName,
		Body:    fmt.Sprintf("您的容器 %s (ID: %s) 已成功启动。", containerName, containerID),
		HTML:    false,
	})
}

func (s *NotificationService) NotifyContainerStopped(ctx context.Context, userEmail, containerID, containerName string) error {
	return s.SendEmail(ctx, &EmailNotification{
		To:      userEmail,
		Subject: "容器已停止 - " + containerName,
		Body:    fmt.Sprintf("您的容器 %s (ID: %s) 已停止。", containerName, containerID),
		HTML:    false,
	})
}

func (s *NotificationService) NotifyGPUHighTemperature(ctx context.Context, adminEmail, serverName string, temperature int) error {
	return s.SendEmail(ctx, &EmailNotification{
		To:      adminEmail,
		Subject: "GPU 高温警告",
		Body:    fmt.Sprintf("服务器 %s 的 GPU 温度已达到 %d°C，请及时处理。", serverName, temperature),
		HTML:    false,
	})
}

func (s *NotificationService) NotifyResourceQuotaWarning(ctx context.Context, userEmail string, used, limit int, resourceType string) error {
	return s.SendEmail(ctx, &EmailNotification{
		To:      userEmail,
		Subject: "资源配额警告",
		Body:    fmt.Sprintf("您的 %s 使用量 (%d) 已接近配额限制 (%d)。请及时清理资源或申请更多配额。", resourceType, used, limit),
		HTML:    false,
	})
}
