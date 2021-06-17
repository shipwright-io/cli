package build

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
	"github.com/shipwright-io/cli/pkg/shp/cmd/buildrun"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/resource"
	"github.com/shipwright-io/cli/test/stub"
)

func delModSetup(t *testing.T) (*buildrun.BuildRunWatcher, *gomega.GomegaWithT, *buildv1alpha1.BuildRun, *resource.Resource) {
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
	brw, err := buildrun.NewBuildRunWatcher(ctx, brr, &metav1.ListOptions{}, &ioStreams)
	g.Expect(err).To(o.BeNil())
	return brw, g, br, brr
}

func addSkipSetup(t *testing.T) (*buildrun.BuildRunWatcher, *gomega.GomegaWithT, *resource.Resource) {
	g := gomega.NewGomegaWithT(t)
	ctx := context.TODO()

	clientset := stub.NewFakeClient()
	p := params.NewParamsForTests(clientset, metav1.NamespaceDefault)
	brr := resource.GetBuildRunResource(p)

	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	brw, err := buildrun.NewBuildRunWatcher(ctx, brr, &metav1.ListOptions{}, &ioStreams)
	g.Expect(err).To(o.BeNil())

	return brw, g, brr
}

func Test_modWatchEvent(t *testing.T) {
	brw, g, br, brr := delModSetup(t)

	modEventCh := make(chan bool, 1)
	eventsDoneCh := make(chan bool, 1)

	brw.WithOnBuildRunModifiedFunc(func(buildRun *buildv1alpha1.BuildRun) (*buildv1alpha1.BuildRun, error) {
		f := getBuildRunWatcherAddModFunc()
		r, e := f(buildRun)
		if r != nil {
			modEventCh <- true
		}
		return r, e
	})

	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := brw.Start()
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	t.Run("buildrun-is-modified", func(t *testing.T) {
		trn := "taskrun"
		br.Status.LatestTaskRunRef = &trn
		err := brr.Update(context.TODO(), br.Name, br)
		g.Expect(err).To(o.BeNil())
	})

	<-modEventCh
	brw.Stop()
	val2, ok := <-eventsDoneCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "event done ch returned val %v and ok %v\n", val2, ok)
	g.Expect(val2).To(o.BeTrue())

}

func Test_delWatchEvent(t *testing.T) {
	brw, g, br, brr := delModSetup(t)

	delEventCh := make(chan bool, 1)
	eventsDoneCh := make(chan bool, 1)

	brw.WithOnBuildRunDeletedFunc(func(buildRun *buildv1alpha1.BuildRun) (*buildv1alpha1.BuildRun, error) {
		f := getBuildRunWatcherDeleteFunc(br.Name)
		r, e := f(buildRun)
		if r != nil {
			delEventCh <- true
		}
		return r, e
	})

	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := brw.Start()
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	t.Run("buildrun-is-deleted", func(t *testing.T) {
		err := brr.Delete(context.TODO(), br.Name)
		g.Expect(err).To(o.BeNil())
	})

	<-delEventCh
	brw.Stop()
	val2, ok := <-eventsDoneCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "event done ch returned val %v and ok %v\n", val2, ok)
	g.Expect(val2).To(o.BeTrue())

}

func Test_skipWatchEvent(t *testing.T) {
	brw, g, brr := addSkipSetup(t)

	skipEventCh := make(chan bool, 1)
	eventsDoneCh := make(chan bool, 1)

	skipBr := &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "skippedbuildrun",
			Namespace: metav1.NamespaceDefault,
		},
	}

	brw.WithSkipBuildRunFunc(func(buildRun *buildv1alpha1.BuildRun) bool {
		f := getBuildRunWatcherSkipFunc("unskippedBuildRun")
		r := f(buildRun)
		if r {
			skipEventCh <- true
		}
		return r
	})

	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := brw.Start()
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	t.Run("buildrun-is-skipped", func(t *testing.T) {
		err := brr.Create(context.TODO(), skipBr.Name, skipBr)
		g.Expect(err).To(o.BeNil())
	})

	<-skipEventCh
	brw.Stop()
	val2, ok := <-eventsDoneCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "event done ch returned val %v and ok %v\n", val2, ok)
	g.Expect(val2).To(o.BeTrue())
}

func Test_addWatchEvent(t *testing.T) {
	brw, g, brr := addSkipSetup(t)

	addEventCh := make(chan bool, 1)
	eventsDoneCh := make(chan bool, 1)

	trn := "taskrun"
	br := &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "buildrun",
			Namespace: metav1.NamespaceDefault,
		},
		Status: buildv1alpha1.BuildRunStatus{LatestTaskRunRef: &trn},
	}
	brw.WithOnBuildRunAddedFunc(func(buildRun *buildv1alpha1.BuildRun) (*buildv1alpha1.BuildRun, error) {
		f := getBuildRunWatcherAddModFunc()
		r, e := f(buildRun)
		if r != nil {
			addEventCh <- true
		}
		return r, e
	})
	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := brw.Start()
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	t.Run("buildrun-is-added", func(t *testing.T) {
		err := brr.Create(context.TODO(), br.Name, br)
		g.Expect(err).To(o.BeNil())
	})

	<-addEventCh
	//<-skipEventCh
	brw.Stop()
	val2, ok := <-eventsDoneCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "event done ch returned val %v and ok %v\n", val2, ok)
	g.Expect(val2).To(o.BeTrue())

}
