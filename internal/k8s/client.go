package k8s

import (
	"context"
	"fmt"
	"time"

	"gpu-platform/internal/config"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sClient struct {
	client    *kubernetes.Clientset
	config    *config.KubernetesConfig
	namespace string
}

func NewK8sClient(cfg *config.KubernetesConfig) (*K8sClient, error) {
	var k8sConfig *rest.Config
	var err error

	if cfg.Kubeconfig != "" {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	} else {
		k8sConfig, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &K8sClient{
		client:    clientset,
		config:    cfg,
		namespace: cfg.Namespace,
	}, nil
}

func (c *K8sClient) CreatePod(ctx context.Context, config *PodConfig) (string, error) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.Name,
			Namespace:   config.Namespace,
			Labels:      config.Labels,
			Annotations: config.Annotations,
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyAlways,
			Containers:    []v1.Container{c.buildContainer(&config.Container)},
			Volumes:       c.buildVolumes(config.Volumes),
			NodeSelector:  config.NodeSelector,
			Tolerations:   c.buildTolerations(config.Tolerations),
		},
	}

	result, err := c.client.CoreV1().Pods(config.Namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create pod: %w", err)
	}

	return result.Name, nil
}

func (c *K8sClient) GetPod(ctx context.Context, name, namespace string) (*v1.Pod, error) {
	pod, err := c.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}
	return pod, nil
}

func (c *K8sClient) DeletePod(ctx context.Context, name, namespace string) error {
	err := c.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}
	return nil
}

func (c *K8sClient) StopPod(ctx context.Context, name, namespace string) error {
	return c.DeletePod(ctx, name, namespace)
}

func (c *K8sClient) WaitForPodReady(ctx context.Context, name, namespace string, timeout time.Duration) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	deadline := time.After(timeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("timeout waiting for pod to be ready")
		case <-ticker.C:
			pod, err := c.GetPod(ctx, name, namespace)
			if err != nil {
				continue
			}

			if isPodReady(pod) {
				return nil
			}
		}
	}
}

func (c *K8sClient) GetPodLogs(ctx context.Context, name, namespace string, tailLines int) ([]string, error) {
	tailLines64 := int64(tailLines)
	logOptions := &v1.PodLogOptions{
		TailLines: &tailLines64,
	}

	req := c.client.CoreV1().Pods(namespace).GetLogs(name, logOptions)
	logs, err := req.Do(ctx).Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %w", err)
	}

	return []string{string(logs)}, nil
}

func (c *K8sClient) ExecCommand(ctx context.Context, name, namespace string, command []string) ([]string, error) {
	return []string{"Command execution disabled"}, nil
}

func (c *K8sClient) GetPodMetrics(ctx context.Context, name, namespace string) (*ContainerMetrics, error) {
	return &ContainerMetrics{}, nil
}

func (c *K8sClient) buildContainer(config *ContainerConfig) v1.Container {
	container := v1.Container{
		Name:            config.Name,
		Image:           config.Image,
		ImagePullPolicy: v1.PullIfNotPresent,
		Command:         config.Command,
		Args:            config.Args,
		Ports:           c.buildContainerPorts(config.Ports),
		Env:             c.buildEnvVars(config.Env),
		Resources:       c.buildResourceRequirements(&config.Resources),
		VolumeMounts:    c.buildVolumeMounts(config.VolumeMounts),
	}

	return container
}

func (c *K8sClient) buildContainerPorts(ports []ContainerPort) []v1.ContainerPort {
	var result []v1.ContainerPort
	for _, p := range ports {
		result = append(result, v1.ContainerPort{
			Name:          p.Name,
			ContainerPort: p.ContainerPort,
			Protocol:      v1.ProtocolTCP,
		})
	}
	return result
}

func (c *K8sClient) buildEnvVars(envs []EnvVar) []v1.EnvVar {
	var result []v1.EnvVar
	for _, e := range envs {
		result = append(result, v1.EnvVar{
			Name:  e.Name,
			Value: e.Value,
		})
	}
	return result
}

func (c *K8sClient) buildResourceRequirements(req *ResourceRequirements) v1.ResourceRequirements {
	resources := v1.ResourceRequirements{
		Limits:   v1.ResourceList{},
		Requests: v1.ResourceList{},
	}

	for name, quantity := range req.Limits {
		qty, _ := resource.ParseQuantity(quantity)
		resources.Limits[v1.ResourceName(name)] = qty
	}

	for name, quantity := range req.Requests {
		qty, _ := resource.ParseQuantity(quantity)
		resources.Requests[v1.ResourceName(name)] = qty
	}

	return resources
}

func (c *K8sClient) buildVolumeMounts(mounts []VolumeMount) []v1.VolumeMount {
	var result []v1.VolumeMount
	for _, m := range mounts {
		result = append(result, v1.VolumeMount{
			Name:      m.Name,
			MountPath: m.MountPath,
		})
	}
	return result
}

func (c *K8sClient) buildVolumes(volumes []Volume) []v1.Volume {
	var result []v1.Volume
	for _, v := range volumes {
		volume := v1.Volume{
			Name: v.Name,
		}

		if v.PersistentVolumeClaim != nil {
			volume.PersistentVolumeClaim = &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: v.PersistentVolumeClaim.ClaimName,
			}
		}

		result = append(result, volume)
	}
	return result
}

func (c *K8sClient) buildTolerations(tolerations []Toleration) []v1.Toleration {
	var result []v1.Toleration
	for _, t := range tolerations {
		result = append(result, v1.Toleration{
			Key:      t.Key,
			Operator: v1.TolerationOperator(t.Operator),
			Value:    t.Value,
			Effect:   v1.TaintEffect(t.Effect),
		})
	}
	return result
}

func isPodReady(pod *v1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

type PodConfig struct {
	Name         string
	Namespace    string
	Labels       map[string]string
	Annotations  map[string]string
	Container    ContainerConfig
	Volumes      []Volume
	NodeSelector map[string]string
	Tolerations  []Toleration
}

type ContainerConfig struct {
	Name         string
	Image        string
	Command      []string
	Args         []string
	Env          []EnvVar
	Resources    ResourceRequirements
	VolumeMounts []VolumeMount
	Ports        []ContainerPort
}

type EnvVar struct {
	Name  string
	Value string
}

type ResourceRequirements struct {
	Limits   ResourceList
	Requests ResourceList
}

type ResourceList map[string]string

type VolumeMount struct {
	Name      string
	MountPath string
}

type ContainerPort struct {
	Name          string
	ContainerPort int32
}

type Volume struct {
	Name                  string
	PersistentVolumeClaim *PersistentVolumeClaimSource
}

type PersistentVolumeClaimSource struct {
	ClaimName string
	SizeGB    int
}

type Toleration struct {
	Key      string
	Operator string
	Value    string
	Effect   string
}

type ContainerMetrics struct {
	CPUUsage      float64
	MemoryUsageMB int
}
