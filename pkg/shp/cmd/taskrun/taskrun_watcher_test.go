package taskrun

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/onsi/gomega"
	o "github.com/onsi/gomega"

	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/resource"
	"github.com/shipwright-io/cli/test/stub"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic/fake"
)

func delModSetup(t *testing.T) (*TaskRunWatcher, *gomega.GomegaWithT, *v1beta1.TaskRun, *resource.Resource) {
	g := gomega.NewGomegaWithT(t)
	ctx := context.TODO()
	scheme := runtime.NewScheme()
	v1beta1.AddToScheme(scheme)
	br := &v1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "taskrun",
			Namespace: metav1.NamespaceDefault,
		},
	}
	dynamicFakeClientset := fake.NewSimpleDynamicClient(scheme, br)
	p := params.NewParamsForTests(dynamicFakeClientset, metav1.NamespaceDefault)
	trr := resource.GetTaskRunResource(p)

	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	trw, err := newTaskRunWatcher(ctx, trr, &metav1.ListOptions{}, &ioStreams)
	g.Expect(err).To(o.BeNil())
	return trw, g, br, trr
}

func addSkipSetup(t *testing.T) (*TaskRunWatcher, *gomega.GomegaWithT, *resource.Resource) {
	g := gomega.NewGomegaWithT(t)
	ctx := context.TODO()

	clientset := stub.NewFakeClient()
	p := params.NewParamsForTests(clientset, metav1.NamespaceDefault)
	trr := resource.GetTaskRunResource(p)

	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	trw, err := newTaskRunWatcher(ctx, trr, &metav1.ListOptions{}, &ioStreams)
	g.Expect(err).To(o.BeNil())

	return trw, g, trr
}

func Test_modWatchEvent(t *testing.T) {
	trw, g, tr, trr := delModSetup(t)

	modEventCh := make(chan bool, 1)
	eventsDoneCh := make(chan bool, 1)

	trw.WithOnTaskRunModifiedFunc(func(taskrun *v1beta1.TaskRun) (*v1beta1.TaskRun, error) {
		f := getTaskRunWatcherAddModFunc()
		r, e := f(taskrun)
		if r != nil {
			modEventCh <- true
		}
		return r, e
	})

	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := trw.Start()
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	t.Run("taskrun-is-modified", func(t *testing.T) {
		tr.Status.PodName = "pod"
		err := trr.Update(context.TODO(), tr.Name, tr)
		g.Expect(err).To(o.BeNil())
	})

	<-modEventCh
	trw.Stop()
	val2, ok := <-eventsDoneCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "event done ch returned val %v and ok %v\n", val2, ok)
	g.Expect(val2).To(o.BeTrue())

}

func Test_delWatchEvent(t *testing.T) {
	trw, g, tr, trr := delModSetup(t)

	delEventCh := make(chan bool, 1)
	eventsDoneCh := make(chan bool, 1)

	trw.WithOnTaskRunDeletedFunc(func(taskrun *v1beta1.TaskRun) (*v1beta1.TaskRun, error) {
		f := getTaskRunWatcherDeleteFunc(tr.Name)
		r, e := f(taskrun)
		if r != nil {
			delEventCh <- true
		}
		return r, e
	})

	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := trw.Start()
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	t.Run("taskrun-is-deleted", func(t *testing.T) {
		err := trr.Delete(context.TODO(), tr.Name)
		g.Expect(err).To(o.BeNil())
	})

	<-delEventCh
	trw.Stop()
	val2, ok := <-eventsDoneCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "event done ch returned val %v and ok %v\n", val2, ok)
	g.Expect(val2).To(o.BeTrue())

}

func Test_skipWatchEvent(t *testing.T) {
	trw, g, trr := addSkipSetup(t)

	skipEventCh := make(chan bool, 1)
	eventsDoneCh := make(chan bool, 1)

	skipBr := &v1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "skippedtaskrun",
			Namespace: metav1.NamespaceDefault,
		},
	}

	trw.WithSkipTaskRunFunc(func(taskrun *v1beta1.TaskRun) bool {
		f := getTaskRunWatcherSkipFunc("unskippedtaskrun")
		r := f(taskrun)
		if r {
			skipEventCh <- true
		}
		return r
	})

	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := trw.Start()
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	t.Run("taskrun-is-skipped", func(t *testing.T) {
		err := trr.Create(context.TODO(), skipBr.Name, skipBr)
		g.Expect(err).To(o.BeNil())
	})

	<-skipEventCh
	trw.Stop()
	val2, ok := <-eventsDoneCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "event done ch returned val %v and ok %v\n", val2, ok)
	g.Expect(val2).To(o.BeTrue())
}

func Test_addWatchEvent(t *testing.T) {
	trw, g, trr := addSkipSetup(t)

	addEventCh := make(chan bool, 1)
	eventsDoneCh := make(chan bool, 1)

	tr := &v1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "taskrun",
			Namespace: metav1.NamespaceDefault,
		},
	}
	tr.Status.PodName = "pod"
	trw.WithOnTaskRunAddedFunc(func(taskrun *v1beta1.TaskRun) (*v1beta1.TaskRun, error) {
		f := getTaskRunWatcherAddModFunc()
		r, e := f(taskrun)
		if r != nil {
			addEventCh <- true
		}
		return r, e
	})
	// executing the event loop in the background, and waiting for the stop channel before inspecting
	// for errors
	go func() {
		_, err := trw.Start()
		g.Expect(err).To(o.BeNil())
		eventsDoneCh <- true
	}()

	t.Run("taskrun-is-added", func(t *testing.T) {
		err := trr.Create(context.TODO(), tr.Name, tr)
		g.Expect(err).To(o.BeNil())
	})

	<-addEventCh
	//<-skipEventCh
	trw.Stop()
	val2, ok := <-eventsDoneCh
	// fyi fmt.Fprintln shows up in real time while t.Logf only shows up after the test completes,
	// so with our multi-threadedness here, that better services live debug
	fmt.Fprintf(os.Stdout, "event done ch returned val %v and ok %v\n", val2, ok)
	g.Expect(val2).To(o.BeTrue())

}
