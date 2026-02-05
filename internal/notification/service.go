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
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	UserID    string                 `json:"userId"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	IsRead    bool                   `json:"isRead"`
	ReadAt    *time.Time             `json:"readAt"`
	CreatedAt time.Time              `json:"createdAt"`
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
		Subject: "[警告] GPU 高温告警",
		Body:    fmt.Sprintf("服务器 %s 的 GPU 温度已达到 %d°C，请及时处理。", serverName, temperature),
		HTML:    false,
	})
}

func (s *NotificationService) NotifyResourceQuotaWarning(ctx context.Context, userEmail string, used, limit int, resourceType string) error {
	percentage := float64(used) / float64(limit) * 100
	return s.SendEmail(ctx, &EmailNotification{
		To:      userEmail,
		Subject: "[警告] 资源配额使用率过高",
		Body:    fmt.Sprintf("您的 %s 使用量已超过 %.1f%% 的配额限制。\n\n当前使用: %d\n配额限制: %d\n\n请及时清理不需要的资源或申请更多配额。", resourceType, percentage, used, limit),
		HTML:    false,
	})
}

func (s *NotificationService) NotifyWelcome(ctx context.Context, userEmail, userName string) error {
	return s.SendEmail(ctx, &EmailNotification{
		To:      userEmail,
		Subject: "欢迎使用 GPU 容器平台",
		Body:    fmt.Sprintf("%s，您好！欢迎加入 GPU 容器平台！", userName),
		HTML:    false,
	})
}
