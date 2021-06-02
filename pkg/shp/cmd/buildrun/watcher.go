package buildrun

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"

	"github.com/shipwright-io/cli/pkg/shp/resource"
)

type BuildRunWatcher struct {
	ctx       context.Context
	stopCh    chan bool
	watcher   watch.Interface
	ioStreams *genericclioptions.IOStreams

	skipBuildRunFunc       SkipBuilRunFunc
	onBuildRunAddedFunc    OnBuildRunEventFunc
	onBuildRunModifiedFunc OnBuildRunEventFunc
	onBuildRunDeletedFunc  OnBuildRunEventFunc
}

type SkipBuilRunFunc func(br *buildv1alpha1.BuildRun) bool

type OnBuildRunEventFunc func(br *buildv1alpha1.BuildRun) (*buildv1alpha1.BuildRun, error)

func (br *BuildRunWatcher) WithSkipBuildRunFunc(fn SkipBuilRunFunc) *BuildRunWatcher {
	br.skipBuildRunFunc = fn
	return br
}

func (br *BuildRunWatcher) WithOnBuildRunAddedFunc(fn OnBuildRunEventFunc) *BuildRunWatcher {
	br.onBuildRunAddedFunc = fn
	return br
}

func (br *BuildRunWatcher) WithOnBuildRunModifiedFunc(fn OnBuildRunEventFunc) *BuildRunWatcher {
	br.onBuildRunModifiedFunc = fn
	return br
}

func (br *BuildRunWatcher) WithOnBuildRunDeletedFunc(fn OnBuildRunEventFunc) *BuildRunWatcher {
	br.onBuildRunDeletedFunc = fn
	return br
}

func (br *BuildRunWatcher) logErrReturnErr(msg string) error {
	fmt.Fprintln(br.ioStreams.ErrOut, msg)
	return fmt.Errorf(msg)
}

func (br *BuildRunWatcher) Start() (*buildv1alpha1.BuildRun, error) {
	exit := false
	for !exit {
		select {
		case <-br.ctx.Done():
			// the shp ' --request-timeout' parameter feeds into the context
			fmt.Fprintf(br.ioStreams.Out, "command context has signalled all current work should exit\n")
			br.watcher.Stop()
			return nil, nil

		case <-br.stopCh:
			fmt.Fprintf(br.ioStreams.Out, "stop on this build run watch has been called\n")
			br.watcher.Stop()
			return nil, nil

		case event := <-br.watcher.ResultChan():
			obj := event.Object
			watchBuildRun := &buildv1alpha1.BuildRun{}
			var err error
			// deletes return runtime.Object vs. unstructured for add/modify with dynamic client
			u, uok := obj.(*unstructured.Unstructured)
			b, bok := obj.(*buildv1alpha1.BuildRun)
			switch {
			case uok:
				err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), watchBuildRun)
				if err != nil {
					return nil, br.logErrReturnErr(fmt.Sprintf("watch error converting Unstructured object to BuildRun: %s\n", err.Error()))
				}
			case bok:
				watchBuildRun = b
			default:
				return nil, br.logErrReturnErr(fmt.Sprintf("watch did not get any expected type, instead got: %#v", obj))
			}

			if br.skipBuildRunFunc != nil && br.skipBuildRunFunc(watchBuildRun) {
				continue
			}

			switch event.Type {
			case watch.Added:
				if br.onBuildRunAddedFunc != nil {
					wbr, err := br.onBuildRunAddedFunc(watchBuildRun)
					if err != nil {
						return nil, err
					}
					if wbr != nil {
						exit = true
					}
					if exit {
						return wbr, nil
					}
				}
			case watch.Modified:
				if br.onBuildRunModifiedFunc != nil {
					wbr, err := br.onBuildRunModifiedFunc(watchBuildRun)
					if err != nil {
						return nil, err
					}
					if wbr != nil {
						exit = true
					}
					if exit {
						return wbr, nil
					}
				}
			case watch.Deleted:
				if br.onBuildRunDeletedFunc != nil {
					wbr, err := br.onBuildRunDeletedFunc(watchBuildRun)
					if err != nil {
						return nil, err
					}
					if wbr != nil {
						exit = true
					}
					if exit {
						return wbr, nil
					}
				}
			}
		}
	}
	return nil, nil
}

func (p *BuildRunWatcher) Stop() {
	close(p.stopCh)
}

func NewBuildRunWatcher(
	ctx context.Context,
	brr *resource.Resource,
	listOpts *metav1.ListOptions,
	ioStreams *genericclioptions.IOStreams) (*BuildRunWatcher, error) {
	brw := &BuildRunWatcher{
		ctx:       ctx,
		ioStreams: ioStreams,
		stopCh:    make(chan bool),
	}
	var err error
	brw.watcher, err = brr.Watch(ctx, *listOpts)
	if err != nil {
		return nil, err
	}

	return brw, nil
}
