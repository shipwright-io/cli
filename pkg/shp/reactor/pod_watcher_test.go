package reactor

import (
	"context"
	"math"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_PodWatcher_RequestTimeout(t *testing.T) {
	g := NewWithT(t)
	ctx := context.TODO()

	clientset := fake.NewSimpleClientset()

	pw := NewPodWatcher(ctx, clientset, metav1.NamespaceDefault, time.Second, 3*time.Second)
	called := false

	pw.WithTimeoutPodFn(func() {
		called = true
	})

	pw.Start(metav1.ListOptions{})
	g.Expect(called).To(BeTrue())
}

func Test_PodWatcher_WatchEventTimeout(t *testing.T) {
	g := NewWithT(t)
	ctx := context.TODO()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      "pod",
		},
	}

	clientset := fake.NewSimpleClientset(pod)

	pw := NewPodWatcher(ctx, clientset, metav1.NamespaceDefault, time.Second, 4*time.Second)
	called := 0

	pw.WithWatchEventTimeoutFn(func(_ *corev1.Pod) (bool, error) {
		called++
		t.Logf("called=%d", called)
		return called > 2, nil
	})

	pw.Start(metav1.ListOptions{})
	g.Expect(called > 0).To(BeTrue())
}

func Test_PodWatcher_ContextTimeout(t *testing.T) {
	g := NewWithT(t)
	ctx := context.TODO()
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second))
	defer cancel()

	clientset := fake.NewSimpleClientset()

	pw := NewPodWatcher(ctxWithDeadline, clientset, metav1.NamespaceDefault, 15*time.Second, math.MaxInt64)
	called := false

	pw.WithTimeoutPodFn(func() {
		called = true
	})

	pw.Start(metav1.ListOptions{})
	g.Expect(called).To(BeTrue())
}

func Test_PodWatcherEvents(t *testing.T) {
	g := NewWithT(t)
	ctx := context.TODO()

	clientset := fake.NewSimpleClientset()

	pw := NewPodWatcher(ctx, clientset, metav1.NamespaceDefault, 15*time.Second, math.MaxInt64)

	eventsCh := make(chan string, 6)
	eventsDoneCh := make(chan bool, 1)

	skipPODFn := "SkipPodFn"
	onPodAddedFn := "OnPodAddedFn"
	onPodDeletedFn := "OnPodDeletedFn"
	onPodModifiedFn := "OnPodModifiedFn"

	// adding functions to be triggered on all types of events, and sending the function name over
	// the events channel
	pw.WithSkipPodFn(func(_ *corev1.Pod) bool {
		eventsCh <- skipPODFn
		return false
	}).WithOnPodAddedFn(func(_ *corev1.Pod) error {
		eventsCh <- onPodAddedFn
		return nil
	}).WithOnPodDeletedFn(func(_ *corev1.Pod) error {
		eventsCh <- onPodDeletedFn
		return nil
	}).WithOnPodModifiedFn(func(_ *corev1.Pod) error {
		eventsCh <- onPodModifiedFn
		return nil
	})

	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		var timeout int64 = 5
		_, err := pw.Start(metav1.ListOptions{
			TimeoutSeconds: &timeout,
		})
		<-pw.stopCh
		g.Expect(err).To(BeNil())
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

	t.Run("pod-is-added", func(_ *testing.T) {
		var err error
		pod, err = podClient.Create(ctx, pod, metav1.CreateOptions{})
		g.Expect(err).To(BeNil())
		time.Sleep(300 * time.Millisecond)
	})

	t.Run("pod-is-modified", func(_ *testing.T) {
		pod.SetLabels(map[string]string{"label": "value"})

		var err error
		pod, err = podClient.Update(ctx, pod, metav1.UpdateOptions{})
		g.Expect(err).To(BeNil())
		time.Sleep(300 * time.Millisecond)
	})

	t.Run("pod-is-deleted", func(_ *testing.T) {
		var gracePeriod int64
		err := podClient.Delete(ctx, pod.GetName(), metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriod,
		})
		g.Expect(err).To(BeNil())
		time.Sleep(300 * time.Millisecond)
	})

	// stopping event-loop running in the background, after waiting for events to arrive on events
	// channel, and before running assertions, it waits for eventsDoneCh as well
	t.Logf("len(eventsCh)=%d", len(eventsCh))
	g.Eventually(len(eventsCh) >= 4, 30*time.Second).Should(BeTrue())

	pw.Stop()
	<-eventsDoneCh
	g.Eventually(pw.stopped).Should(BeTrue())

	// asserting that all events have been exercised, by inspecting the function names sent over the
	// events channel
	g.Eventually(eventsCh).Should(Receive(&skipPODFn))
	g.Eventually(eventsCh).Should(Receive(&onPodAddedFn))
	g.Eventually(eventsCh).Should(Receive(&onPodModifiedFn))
	g.Eventually(eventsCh).Should(Receive(&onPodDeletedFn))
}
