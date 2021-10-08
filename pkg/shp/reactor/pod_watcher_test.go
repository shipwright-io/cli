package reactor

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
	o "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_PodWatcherEvents(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ctx := context.TODO()

	clientset := fake.NewSimpleClientset()

	pw, err := NewPodWatcher(ctx, clientset, metav1.ListOptions{}, metav1.NamespaceDefault)
	g.Expect(err).To(o.BeNil())

	eventsCh := make(chan string, 5)
	eventsDoneCh := make(chan bool, 1)

	skipPODFn := "SkipPodFn"
	onPodAddedFn := "OnPodAddedFn"
	onPodDeletedFn := "OnPodDeletedFn"
	onPodModifiedFn := "OnPodModifiedFn"

	// adding functions to be triggered on all types of events, and sending the function name over
	// the events channel
	pw.WithSkipPodFn(func(pod *corev1.Pod) bool {
		eventsCh <- skipPODFn
		return false
	}).WithOnPodAddedFn(func(pod *corev1.Pod) error {
		eventsCh <- onPodAddedFn
		return nil
	}).WithOnPodDeletedFn(func(pod *corev1.Pod) error {
		eventsCh <- onPodDeletedFn
		return nil
	}).WithOnPodModifiedFn(func(pod *corev1.Pod) error {
		eventsCh <- onPodModifiedFn
		return nil
	})

	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := pw.Start()
		<-pw.stopCh
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      "pod",
		},
	}

	// making modifications in the pod, making sure all events are exercised, thus the events channel
	// should be populated
	podClient := clientset.CoreV1().Pods(metav1.NamespaceDefault)

	t.Run("pod-is-added", func(t *testing.T) {
		var err error
		pod, err = podClient.Create(ctx, pod, metav1.CreateOptions{})
		g.Expect(err).To(o.BeNil())
	})

	t.Run("pod-is-modified", func(t *testing.T) {
		pod.SetLabels(map[string]string{"label": "value"})

		var err error
		pod, err = podClient.Update(ctx, pod, metav1.UpdateOptions{})
		g.Expect(err).To(o.BeNil())
	})

	t.Run("pod-is-deleted", func(t *testing.T) {
		err := podClient.Delete(ctx, pod.GetName(), metav1.DeleteOptions{})
		g.Expect(err).To(o.BeNil())
	})

	// stopping event-loop running in the background, after waiting for events to arrive on events
	// channel, and before running assertions, it waits for eventsDoneCh as well
	<-eventsCh
	pw.Stop()
	<-eventsDoneCh

	// asserting that all events have been exercised, by inspecting the function names sent over the
	// events channel
	g.Eventually(eventsCh).Should(o.Receive(&skipPODFn))
	g.Eventually(eventsCh).Should(o.Receive(&onPodAddedFn))
	g.Eventually(eventsCh).Should(o.Receive(&onPodModifiedFn))
	g.Eventually(eventsCh).Should(o.Receive(&onPodDeletedFn))
}
