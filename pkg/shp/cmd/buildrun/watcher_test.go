package buildrun

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/onsi/gomega"
	o "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic/fake"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/resource"
	"github.com/shipwright-io/cli/test/stub"
)

func Test_BuildRunWatcherCreateAndSkipEvent(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ctx := context.TODO()

	clientset := stub.NewFakeClient()
	p := params.NewParamsForTests(clientset, metav1.NamespaceDefault)
	brr := resource.GetBuildRunResource(p)

	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	brw, err := NewBuildRunWatcher(ctx, brr, &metav1.ListOptions{}, &ioStreams)
	g.Expect(err).To(o.BeNil())

	addEventCh := make(chan string, 1)
	skipEventCh := make(chan string, 1)
	eventsDoneCh := make(chan bool, 1)
	onBuildRunAddedFn := "OnBuildRunAddedFn"
	skipBuildRunFn := "SkipBuildRunFn"

	brw.WithOnBuildRunAddedFunc(func(br *buildv1alpha1.BuildRun) (*buildv1alpha1.BuildRun, error) {
		// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
		// so with our multi-threadedness here, that better services live debug
		fmt.Fprintln(os.Stdout, onBuildRunAddedFn+" called")
		addEventCh <- onBuildRunAddedFn
		return nil, nil
	}).WithSkipBuildRunFunc(func(br *buildv1alpha1.BuildRun) bool {
		// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
		// so with our multi-threadedness here, that better services live debug
		fmt.Fprintln(os.Stdout, skipBuildRunFn+" called")
		skipEventCh <- skipBuildRunFn
		return false
	})

	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := brw.Start()
		<-brw.stopCh
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	br := &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "buildrun",
			Namespace: metav1.NamespaceDefault,
		},
	}

	t.Run("buildrun-is-added", func(t *testing.T) {
		err := brr.Create(context.TODO(), br.Name, br)
		g.Expect(err).To(o.BeNil())
	})

	val, ok := <-addEventCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "add event ch returned val %v and ok %v\n", val, ok)
	g.Expect(val).To(o.Equal(onBuildRunAddedFn))
	val, ok = <-skipEventCh
	fmt.Fprintf(os.Stdout, "skip event ch returned val %v and ok %v\n", val, ok)
	g.Expect(val).To(o.Equal(skipBuildRunFn))

	brw.Stop()
	val2, ok := <-eventsDoneCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "event done ch returned val %v and ok %v\n", val2, ok)
	g.Expect(val2).To(o.BeTrue())

}

func delModSetup(t *testing.T) (*BuildRunWatcher, *gomega.GomegaWithT, *buildv1alpha1.BuildRun, *resource.Resource) {
	g := gomega.NewGomegaWithT(t)
	ctx := context.TODO()
	scheme := runtime.NewScheme()
	buildv1alpha1.SchemeBuilder.AddToScheme(scheme)
	br := &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "buildrun",
			Namespace: metav1.NamespaceDefault,
		},
	}
	dynamicFakeClientset := fake.NewSimpleDynamicClient(scheme, br)
	p := params.NewParamsForTests(dynamicFakeClientset, metav1.NamespaceDefault)
	brr := resource.GetBuildRunResource(p)

	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	brw, err := NewBuildRunWatcher(ctx, brr, &metav1.ListOptions{}, &ioStreams)
	g.Expect(err).To(o.BeNil())
	return brw, g, br, brr
}

func Test_BuildRunWatcherModifyEvent(t *testing.T) {
	brw, g, br, brr := delModSetup(t)

	eventsDoneCh := make(chan bool, 1)
	modEventCh := make(chan string, 1)
	onBuildRunModifiedFn := "OnBuildRunModifiedFn"

	brw.WithOnBuildRunModifiedFunc(func(br *buildv1alpha1.BuildRun) (*buildv1alpha1.BuildRun, error) {
		// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
		// so with our multi-threadedness here, that better services live debug
		fmt.Fprintln(os.Stdout, onBuildRunModifiedFn+" called")
		modEventCh <- onBuildRunModifiedFn
		return nil, nil
	})

	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := brw.Start()
		<-brw.stopCh
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	t.Run("pod-is-modified", func(t *testing.T) {
		br.SetLabels(map[string]string{"label": "value"})

		err := brr.Update(context.TODO(), br.Name, br)
		g.Expect(err).To(o.BeNil())
	})

	val, ok := <-modEventCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "mod event ch returned val %v and ok %v\n", val, ok)

	brw.Stop()
	val2, ok := <-eventsDoneCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "event done ch returned val %v and ok %v\n", val2, ok)
	g.Expect(val2).To(o.BeTrue())
}

func Test_BuildRunWatcherDeleteEvent(t *testing.T) {
	brw, g, br, brr := delModSetup(t)

	delEventCh := make(chan string, 1)
	eventsDoneCh := make(chan bool, 1)
	onBuildRunDeletedFn := "OnBuildRunDeletedFn"

	// adding functions to be triggered on all types of events, and sending the function name over
	// the events channel
	brw.WithOnBuildRunDeletedFunc(func(br *buildv1alpha1.BuildRun) (*buildv1alpha1.BuildRun, error) {
		// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
		// so with our multi-threadedness here, that better services live debug
		fmt.Fprintln(os.Stdout, onBuildRunDeletedFn+" called")
		delEventCh <- onBuildRunDeletedFn
		return nil, nil
	})

	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := brw.Start()
		<-brw.stopCh
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	t.Run("pod-is-deleted", func(t *testing.T) {
		err := brr.Delete(context.TODO(), br.Name)
		g.Expect(err).To(o.BeNil())
	})

	val, ok := <-delEventCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "del event ch returned val %v and ok %v\n", val, ok)

	brw.Stop()
	val2, ok := <-eventsDoneCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "event done ch returned val %v and ok %v\n", val2, ok)
	g.Expect(val2).To(o.BeTrue())
}
