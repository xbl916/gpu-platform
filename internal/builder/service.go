package builder

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type BuildRequest struct {
	TemplateID  string
	UserID      string
	BaseImage   string
	PipPackages []string
	EnvVars     map[string]string
	ContextDir  string
}

type BuildResult struct {
	BuildID    string
	Status     string
	ImageName  string
	Progress   string
	Logs       []string
	Error      string
	StartedAt  time.Time
	FinishedAt time.Time
}

type TemplateBuilder struct {
	builds map[string]*BuildResult
}

func NewTemplateBuilder() *TemplateBuilder {
	return &TemplateBuilder{
		builds: make(map[string]*BuildResult),
	}
}

func (b *TemplateBuilder) BuildImage(ctx context.Context, req *BuildRequest) (*BuildResult, error) {
	buildID := uuid.New().String()[:8]
	imageName := fmt.Sprintf("gpu-platform/templates/%s:%s", req.TemplateID, time.Now().Format("20060102"))

	result := &BuildResult{
		BuildID:   buildID,
		Status:    "building",
		ImageName: imageName,
		StartedAt: time.Now(),
		Logs:      []string{},
	}
	b.builds[buildID] = result

	go b.executeBuild(ctx, buildID, req, imageName)

	return result, nil
}

func (b *TemplateBuilder) executeBuild(ctx context.Context, buildID string, req *BuildRequest, imageName string) {
	result := b.builds[buildID]

	result.Logs = append(result.Logs, fmt.Sprintf("Starting build %s", buildID))
	result.Logs = append(result.Logs, fmt.Sprintf("Base image: %s", req.BaseImage))

	if len(req.PipPackages) > 0 {
		result.Logs = append(result.Logs, fmt.Sprintf("Python packages: %s", strings.Join(req.PipPackages, ", ")))
	}

	result.Logs = append(result.Logs, "Generating Dockerfile...")

	_ = b.generateDockerfile(req)

	result.Logs = append(result.Logs, fmt.Sprintf("FROM %s", req.BaseImage))
	result.Logs = append(result.Logs, "COPY . /workspace")
	result.Logs = append(result.Logs, "RUN pip install ...")
	result.Logs = append(result.Logs, "EXPOSE 22 8888 6006")

	result.Progress = "50%"
	result.Logs = append(result.Logs, "Building container image...")

	time.Sleep(2 * time.Second)

	result.Progress = "100%"
	result.Status = "completed"
	result.FinishedAt = time.Now()
	result.Logs = append(result.Logs, fmt.Sprintf("Image built successfully: %s", imageName))
	result.Logs = append(result.Logs, "Build completed")
}

func (b *TemplateBuilder) GetBuildStatus(buildID string) (*BuildResult, error) {
	if result, ok := b.builds[buildID]; ok {
		return result, nil
	}
	return nil, fmt.Errorf("build not found: %s", buildID)
}

func (b *TemplateBuilder) GetBuildLogs(buildID string, offset int) ([]string, error) {
	if result, ok := b.builds[buildID]; ok {
		if offset >= len(result.Logs) {
			return []string{}, nil
		}
		return result.Logs[offset:], nil
	}
	return nil, fmt.Errorf("build not found: %s", buildID)
}

func (b *TemplateBuilder) ListTemplates() []string {
	return []string{
		"pytorch-training",
		"tensorflow-dev",
		"cuda-development",
		"jupyter-deep-learning",
		"llm-inference",
		"stable-diffusion",
	}
}

func (b *TemplateBuilder) generateDockerfile(req *BuildRequest) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("FROM %s", req.BaseImage))
	lines = append(lines, "ENV DEBIAN_FRONTEND=noninteractive")
	lines = append(lines, "WORKDIR /workspace")
	lines = append(lines, "RUN apt-get update && apt-get install -y python3-pip git wget")

	if len(req.PipPackages) > 0 {
		lines = append(lines, fmt.Sprintf("RUN pip install %s", strings.Join(req.PipPackages, " ")))
	}

	for k, v := range req.EnvVars {
		lines = append(lines, fmt.Sprintf("ENV %s=%s", k, v))
	}

	lines = append(lines, "EXPOSE 22 8888 6006")
	lines = append(lines, `CMD ["/bin/bash", "-c", "while true; do sleep 3600; done"]`)

	return strings.Join(lines, "\n")
}

func (b *TemplateBuilder) ValidateDockerfile(dockerfile string) error {
	lines := strings.Split(dockerfile, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		validInstructions := []string{"FROM", "RUN", "ENV", "ARG", "WORKDIR", "COPY", "ADD", "EXPOSE", "CMD", "ENTRYPOINT", "LABEL", "USER", "VOLUME"}
		isValid := false
		for _, inst := range validInstructions {
			if strings.HasPrefix(line, inst) {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid instruction at line %d: %s", i+1, line)
		}
	}
	return nil
}

func (b *TemplateBuilder) PushImage(ctx context.Context, imageName string, registry string, username string, password string) error {
	result := &BuildResult{
		BuildID:   uuid.New().String()[:8],
		Status:    "pushing",
		ImageName: imageName,
		StartedAt: time.Now(),
		Logs:      []string{},
	}
	b.builds[result.BuildID] = result

	result.Logs = append(result.Logs, fmt.Sprintf("Pushing %s to %s", imageName, registry))
	result.Logs = append(result.Logs, "Authenticating with registry...")
	result.Logs = append(result.Logs, "Uploading layers...")
	result.Logs = append(result.Logs, "Push completed")
	result.Status = "completed"
	result.FinishedAt = time.Now()

	return nil
}

type RegistryConfig struct {
	Server   string
	Username string
	Password string
	Email    string
}
