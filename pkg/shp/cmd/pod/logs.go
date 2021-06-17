package pod

import (
	"bytes"
	"context"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPodLogs returns log output of the k8s container provided by pod and name
func GetPodLogs(ctx context.Context, client kubernetes.Interface, pod corev1.Pod, container string) (string, error) {
	podLogOpts := corev1.PodLogOptions{
		Container: container,
	}
	req := client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
