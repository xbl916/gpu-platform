package builder

import "context"

type BuildRequest struct {
	TemplateID  string
	UserID      string
	BaseImage   string
	PipPackages []string
	EnvVars     map[string]string
}

type BuildResult struct {
	BuildID   string
	Status    string
	ImageName string
	Error     string
}

type TemplateBuilder struct{}

func NewTemplateBuilder() *TemplateBuilder {
	return &TemplateBuilder{}
}

func (b *TemplateBuilder) BuildImage(ctx context.Context, req *BuildRequest) (*BuildResult, error) {
	return &BuildResult{
		BuildID:   "demo-build-id",
		Status:    "building",
		ImageName: "gpu-platform/templates/demo:latest",
	}, nil
}

func (b *TemplateBuilder) GetBuildStatus(buildID string) (*BuildResult, error) {
	return &BuildResult{
		BuildID:   buildID,
		Status:    "completed",
		ImageName: "gpu-platform/templates/demo:latest",
	}, nil
}

func (b *TemplateBuilder) ListTemplates() []string {
	return []string{"pytorch-training", "tensorflow-dev", "cuda-development"}
}
