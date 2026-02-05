package terminal

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"gpu-platform/internal/config"
	"gpu-platform/internal/k8s"
)

type TerminalService struct {
	cfg       *config.Config
	k8sClient *k8s.K8sClient
	sessions  map[string]*TerminalSession
	mu        sync.RWMutex
}

type TerminalSession struct {
	ID        string
	PodName   string
	UserID    string
	Cols      int
	Rows      int
	CreatedAt time.Time
}

func NewTerminalService(cfg *config.Config, k8sClient *k8s.K8sClient) *TerminalService {
	return &TerminalService{
		cfg:       cfg,
		k8sClient: k8sClient,
		sessions:  make(map[string]*TerminalSession),
	}
}

func (s *TerminalService) ExecuteCommand(ctx context.Context, podName string, command string) ([]string, error) {
	return s.k8sClient.ExecCommand(ctx, podName, []string{"/bin/bash", "-c", command})
}

func (s *TerminalService) CreateSession(podName, userID string) *TerminalSession {
	session := &TerminalSession{
		ID:        fmt.Sprintf("term-%d", time.Now().UnixNano()),
		PodName:   podName,
		UserID:    userID,
		Cols:      80,
		Rows:      24,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()

	return session
}

func (s *TerminalService) ResizeSession(sessionID string, cols, rows int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.sessions[sessionID]; ok {
		session.Cols = cols
		session.Rows = rows
		return nil
	}
	return fmt.Errorf("session not found")
}

func (s *TerminalService) GetSession(sessionID string) (*TerminalSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if session, ok := s.sessions[sessionID]; ok {
		return session, nil
	}
	return nil, fmt.Errorf("session not found")
}

func (s *TerminalService) CloseSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

func (s *TerminalService) GetActiveSessions(userID string) []TerminalSessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sessions []TerminalSessionInfo
	for _, session := range s.sessions {
		if session.UserID == userID {
			sessions = append(sessions, TerminalSessionInfo{
				ID:        session.ID,
				PodName:   session.PodName,
				CreatedAt: session.CreatedAt,
			})
		}
	}
	return sessions
}

type TerminalSessionInfo struct {
	ID        string
	PodName   string
	CreatedAt time.Time
}

type SSHService struct {
	cfg *config.Config
}

func NewSSHService(cfg *config.Config) *SSHService {
	return &SSHService{cfg: cfg}
}

func (s *SSHService) GetSSHConfig(containerID string, userID string) (*SSHConfig, error) {
	return &SSHConfig{
		Host:     fmt.Sprintf("ssh.gpu-platform.local"),
		Port:     22,
		Username: userID,
		Password: generatePassword(),
	}, nil
}

func generatePassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 16
	var password strings.Builder
	for i := 0; i < length; i++ {
		password.WriteByte(charset[i%len(charset)])
	}
	return password.String()
}

type SSHConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type WebAccessService struct {
	cfg       *config.Config
	k8sClient *k8s.K8sClient
}

func NewWebAccessService(cfg *config.Config, k8sClient *k8s.K8sClient) *WebAccessService {
	return &WebAccessService{
		cfg:       cfg,
		k8sClient: k8sClient,
	}
}

func (s *WebAccessService) GetAccessInfo(ctx context.Context, containerID string) (*AccessInfo, error) {
	pod, err := s.k8sClient.GetPod(ctx, containerID)
	if err != nil {
		return nil, err
	}

	return &AccessInfo{
		ContainerID: containerID,
		Status:      pod.Status,
		IP:          pod.IP,
		SSH: SSHEndpoint{
			Host: pod.IP,
			Port: 22,
		},
		Jupyter: WebEndpoint{
			URL:  fmt.Sprintf("http://%s:8888", pod.IP),
			Port: 8888,
		},
		TensorBoard: WebEndpoint{
			URL:  fmt.Sprintf("http://%s:6006", pod.IP),
			Port: 6006,
		},
	}, nil
}

func (s *WebAccessService) CreateIngress(ctx context.Context, name string, host string, serviceName string, servicePort int) error {
	return nil
}

type AccessInfo struct {
	ContainerID string
	Status      string
	IP          string
	SSH         SSHEndpoint
	Jupyter     WebEndpoint
	TensorBoard WebEndpoint
}

type SSHEndpoint struct {
	Host string
	Port int
}

type WebEndpoint struct {
	URL  string
	Port int
}
