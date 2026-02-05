package k8s

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"gpu-platform/internal/config"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type K8sClient struct {
	client    *kubernetes.Clientset
	namespace string
	config    *rest.Config
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
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &K8sClient{
		client:    clientset,
		namespace: cfg.Namespace,
		config:    k8sConfig,
	}, nil
}

func (c *K8sClient) IsReady(ctx context.Context) error {
	_, err := c.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	return err
}

func (c *K8sClient) CreatePodWithPVC(ctx context.Context, name string, image string, gpuCount int, resources ResourceRequirements, storageGB int, envVars map[string]string) (*PodInfo, error) {
	gpuResource := fmt.Sprintf("%d", gpuCount)

	env := []v1.EnvVar{
		{Name: "NVIDIA_VISIBLE_DEVICES", Value: "all"},
		{Name: "NVIDIA_DRIVER_CAPABILITIES", Value: "compute,utility"},
		{Name: "CUDA_VISIBLE_DEVICES", Value: "0,1,2,3"},
		{Name: "USER_ID", Value: "user"},
	}

	for k, v := range envVars {
		env = append(env, v1.EnvVar{Name: k, Value: v})
	}

	pvcName := fmt.Sprintf("data-%s", name)
	workspacePVCName := fmt.Sprintf("workspace-%s", name)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.namespace,
			Labels: map[string]string{
				"app":                       name,
				"gpu-platform.io/owner":     "user",
				"gpu-platform.io/container": name,
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyAlways,
			NodeSelector: map[string]string{
				"gpu-platform.io/gpu": "true",
			},
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{Key: "nvidia.com/gpu.count", Operator: v1.NodeSelectorOpExists},
								},
							},
						},
					},
				},
			},
			Containers: []v1.Container{
				{
					Name:            "main",
					Image:           image,
					ImagePullPolicy: v1.PullIfNotPresent,
					Command:         []string{"/bin/bash", "-c"},
					Args:            []string{"trap exit 0 SIGTERM; while true; do sleep 3600; done"},
					Env:             env,
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"nvidia.com/gpu": resource.MustParse(gpuResource),
							"cpu":            resource.MustParse(fmt.Sprintf("%d", resources.CPU)),
							"memory":         resource.MustParse(fmt.Sprintf("%dMi", resources.Memory)),
						},
						Requests: v1.ResourceList{
							"nvidia.com/gpu": resource.MustParse(gpuResource),
							"cpu":            resource.MustParse(fmt.Sprintf("%d", resources.CPU)),
							"memory":         resource.MustParse(fmt.Sprintf("%dMi", resources.Memory)),
						},
					},
					Ports: []v1.ContainerPort{
						{ContainerPort: 22, Name: "ssh", Protocol: v1.ProtocolTCP},
						{ContainerPort: 8888, Name: "jupyter", Protocol: v1.ProtocolTCP},
						{ContainerPort: 6006, Name: "tensorboard", Protocol: v1.ProtocolTCP},
					},
					VolumeMounts: []v1.VolumeMount{
						{Name: "workspace", MountPath: "/workspace"},
						{Name: "data", MountPath: "/data"},
					},
					LivenessProbe: &v1.Probe{
						InitialDelaySeconds: 30,
						PeriodSeconds:       10,
						TimeoutSeconds:      5,
						FailureThreshold:    3,
						ProbeHandler: v1.ProbeHandler{
							Exec: &v1.ExecAction{
								Command: []string{"sh", "-c", "ps aux | grep -v grep | grep -q 'sleep 3600'"},
							},
						},
					},
					ReadinessProbe: &v1.Probe{
						InitialDelaySeconds: 10,
						PeriodSeconds:       5,
						TimeoutSeconds:      3,
						FailureThreshold:    3,
						ProbeHandler: v1.ProbeHandler{
							Exec: &v1.ExecAction{
								Command: []string{"sh", "-c", "ps aux | grep -v grep | grep -q 'sleep 3600'"},
							},
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "workspace",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: workspacePVCName,
						},
					},
				},
				{
					Name: "data",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
						},
					},
				},
			},
			Tolerations: []v1.Toleration{
				{Key: "nvidia.com/gpu", Operator: v1.TolerationOpExists, Effect: v1.TaintEffectNoSchedule},
				{Key: "gpu", Operator: v1.TolerationOpExists, Effect: v1.TaintEffectNoSchedule},
			},
			TerminationGracePeriodSeconds: new(int64),
		},
	}

	if storageGB > 0 {
		if err := c.CreatePVC(ctx, workspacePVCName, 50); err != nil {
			return nil, fmt.Errorf("failed to create workspace PVC: %w", err)
		}
		if err := c.CreatePVC(ctx, pvcName, storageGB); err != nil {
			return nil, fmt.Errorf("failed to create data PVC: %w", err)
		}
	}

	result, err := c.client.CoreV1().Pods(c.namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %w", err)
	}

	return &PodInfo{
		Name:      result.Name,
		Namespace: result.Namespace,
		IP:        result.Status.PodIP,
	}, nil
}

func (c *K8sClient) CreatePod(ctx context.Context, name string, image string, gpuCount int, resources ResourceRequirements) (*PodInfo, error) {
	return c.CreatePodWithPVC(ctx, name, image, gpuCount, resources, 0, nil)
}

func (c *K8sClient) GetPod(ctx context.Context, name string) (*PodInfo, error) {
	pod, err := c.client.CoreV1().Pods(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	return &PodInfo{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		IP:        pod.Status.PodIP,
		Status:    string(pod.Status.Phase),
	}, nil
}

func (c *K8sClient) DeletePod(ctx context.Context, name string) error {
	err := c.client.CoreV1().Pods(c.namespace).Delete(ctx, name, metav1.DeleteOptions{
		GracePeriodSeconds: new(int64),
	})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}
	return nil
}

func (c *K8sClient) WaitForReady(ctx context.Context, name string, timeout time.Duration) error {
	return wait.PollImmediate(5*time.Second, timeout, func() (bool, error) {
		pod, err := c.client.CoreV1().Pods(c.namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
}

func (c *K8sClient) WaitForTerminated(ctx context.Context, name string, timeout time.Duration) error {
	return wait.PollImmediate(2*time.Second, timeout, func() (bool, error) {
		_, err := c.client.CoreV1().Pods(c.namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}

func (c *K8sClient) StartPod(ctx context.Context, name string) error {
	pod, err := c.client.CoreV1().Pods(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	maxDuration := int64(0)
	pod.Spec.ActiveDeadlineSeconds = &maxDuration

	_, err = c.client.CoreV1().Pods(c.namespace).Update(ctx, pod, metav1.UpdateOptions{})
	return err
}

func (c *K8sClient) StopPod(ctx context.Context, name string) error {
	maxDuration := int64(300)
	err := c.client.CoreV1().Pods(c.namespace).Delete(ctx, name, metav1.DeleteOptions{
		GracePeriodSeconds: &maxDuration,
	})
	return err
}

func (c *K8sClient) GetPodLogs(ctx context.Context, name string, tailLines int64) ([]string, error) {
	logOptions := &v1.PodLogOptions{TailLines: &tailLines}
	req := c.client.CoreV1().Pods(c.namespace).GetLogs(name, logOptions)

	logs, err := req.Do(ctx).Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(logs))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

func (c *K8sClient) ExecCommand(ctx context.Context, podName string, command []string) ([]string, error) {
	req := c.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(c.namespace).
		SubResource("exec")

	option := &v1.PodExecOptions{
		Command: []string{"/bin/sh", "-c"},
		Stdout:  true,
		Stderr:  true,
	}

	req.VersionedParams(option, metav1.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	var output []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line != "" {
			output = append(output, line)
		}
	}

	if stderr.Len() > 0 {
		output = append(output, "STDERR: "+stderr.String())
	}

	return output, err
}

func (c *K8sClient) CreatePVC(ctx context.Context, name string, sizeGi int) error {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.namespace,
			Labels: map[string]string{
				"gpu-platform.io/resource": "storage",
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dGi", sizeGi)),
				},
			},
		},
	}

	_, err := c.client.CoreV1().PersistentVolumeClaims(c.namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create pvc: %w", err)
	}

	return nil
}

func (c *K8sClient) DeletePVC(ctx context.Context, name string) error {
	err := c.client.CoreV1().PersistentVolumeClaims(c.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pvc: %w", err)
	}
	return nil
}

func (c *K8sClient) ListPods(ctx context.Context, labelSelector string) ([]PodInfo, error) {
	pods, err := c.client.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var result []PodInfo
	for _, pod := range pods.Items {
		result = append(result, PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			IP:        pod.Status.PodIP,
			Status:    string(pod.Status.Phase),
		})
	}

	return result, nil
}

func (c *K8sClient) GetPodMetrics(ctx context.Context, podName string) (*PodMetrics, error) {
	pod, err := c.client.CoreV1().Pods(c.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	metrics := &PodMetrics{
		Name:  pod.Name,
		Phase: string(pod.Status.Phase),
		IP:    pod.Status.PodIP,
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == "main" {
			metrics.RestartCount = int(containerStatus.RestartCount)
			if containerStatus.State.Running != nil {
				metrics.State = "running"
				metrics.StartedAt = containerStatus.State.Running.StartedAt.Time
			} else if containerStatus.State.Waiting != nil {
				metrics.State = "waiting"
				metrics.WaitingReason = containerStatus.State.Waiting.Reason
			} else if containerStatus.State.Terminated != nil {
				metrics.State = "terminated"
			}
		}
	}

	return metrics, nil
}

func (c *K8sClient) CreateService(ctx context.Context, name string, port int) error {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []v1.ServicePort{
				{
					Name:     "ssh",
					Protocol: v1.ProtocolTCP,
					Port:     int32(port),
					NodePort: int32(port + 30000),
				},
			},
			Type: v1.ServiceTypeNodePort,
		},
	}

	_, err := c.client.CoreV1().Services(c.namespace).Create(ctx, svc, metav1.CreateOptions{})
	return err
}

type PodInfo struct {
	Name      string
	Namespace string
	IP        string
	Status    string
}

type ResourceRequirements struct {
	CPU    int
	Memory int
}

type PodMetrics struct {
	Name          string
	Phase         string
	IP            string
	State         string
	RestartCount  int
	WaitingReason string
	StartedAt     time.Time
	CPUUsage      float64
	MemoryUsage   float64
}
