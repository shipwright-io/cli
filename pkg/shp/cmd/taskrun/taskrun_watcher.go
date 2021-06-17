package taskrun

import (
	"context"
	"fmt"

	"github.com/shipwright-io/cli/pkg/shp/resource"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func getTaskRunWatcherAddModFunc() func(tr *v1beta1.TaskRun) (*v1beta1.TaskRun, error) {
	f := func(tr *v1beta1.TaskRun) (*v1beta1.TaskRun, error) {
		if tr.IsDone() {
			return tr, nil
		}
		if len(tr.Status.PodName) > 0 {
			return tr, nil
		}
		return nil, nil
	}
	return f
}

func getTaskRunWatcherSkipFunc(taskRunName string) func(tr *v1beta1.TaskRun) bool {
	return func(tr *v1beta1.TaskRun) bool {
		if tr.Name != taskRunName {
			return true
		}
		return false
	}
}

func getTaskRunWatcherDeleteFunc(taskRunName string) func(tr *v1beta1.TaskRun) (*v1beta1.TaskRun, error) {
	return func(tr *v1beta1.TaskRun) (*v1beta1.TaskRun, error) {
		return tr, nil
	}
}

func WaitForTaskRunToHavePod(ctx context.Context, taskRunName string, trr *resource.Resource, ioStreams *genericclioptions.IOStreams) (*v1beta1.TaskRun, error) {
	trw, err := newTaskRunWatcher(ctx, trr, &metav1.ListOptions{}, ioStreams)
	if err != nil {
		return nil, err
	}
	f := getTaskRunWatcherAddModFunc()
	trw.WithSkipTaskRunFunc(getTaskRunWatcherSkipFunc(taskRunName)).
		WithOnTaskRunDeletedFunc(getTaskRunWatcherDeleteFunc(taskRunName)).
		WithOnTaskRunAddedFunc(f).
		WithOnTaskRunModifiedFunc(f)
	return trw.Start()
}

type TaskRunWatcher struct {
	ctx       context.Context
	stopCh    chan bool
	watcher   watch.Interface
	ioStreams *genericclioptions.IOStreams

	skipTaskRunFunc       SkipTaskRunFunc
	onTaskRunAddedFunc    OnTaskRunEventFunc
	onTaskRunModifiedFunc OnTaskRunEventFunc
	onTaskRunDeleteFunc   OnTaskRunEventFunc
}

type SkipTaskRunFunc func(tr *v1beta1.TaskRun) bool

type OnTaskRunEventFunc func(tr *v1beta1.TaskRun) (*v1beta1.TaskRun, error)

func (tr *TaskRunWatcher) WithSkipTaskRunFunc(fn SkipTaskRunFunc) *TaskRunWatcher {
	tr.skipTaskRunFunc = fn
	return tr
}

func (tr *TaskRunWatcher) WithOnTaskRunAddedFunc(fn OnTaskRunEventFunc) *TaskRunWatcher {
	tr.onTaskRunAddedFunc = fn
	return tr
}

func (tr *TaskRunWatcher) WithOnTaskRunModifiedFunc(fn OnTaskRunEventFunc) *TaskRunWatcher {
	tr.onTaskRunModifiedFunc = fn
	return tr
}

func (tr *TaskRunWatcher) WithOnTaskRunDeletedFunc(fn OnTaskRunEventFunc) *TaskRunWatcher {
	tr.onTaskRunDeleteFunc = fn
	return tr
}

func (tr *TaskRunWatcher) logErrReturnErr(msg string) error {
	fmt.Fprintln(tr.ioStreams.ErrOut, msg)
	return fmt.Errorf(msg)
}

func (br *TaskRunWatcher) Start() (*v1beta1.TaskRun, error) {
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
			watchTaskRun := &v1beta1.TaskRun{}
			var err error
			// deletes return runtime.Object vs. unstructured for add/modify with dynamic client
			u, uok := obj.(*unstructured.Unstructured)
			b, bok := obj.(*v1beta1.TaskRun)
			switch {
			case uok:
				err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), watchTaskRun)
				if err != nil {
					return nil, br.logErrReturnErr(fmt.Sprintf("watch error converting Unstructured object to BuildRun: %s\n", err.Error()))
				}
			case bok:
				watchTaskRun = b
			default:
				return nil, br.logErrReturnErr(fmt.Sprintf("watch did not get any expected type, instead got: %#v", obj))
			}

			if br.skipTaskRunFunc != nil && br.skipTaskRunFunc(watchTaskRun) {
				continue
			}

			switch event.Type {
			case watch.Added:
				if br.onTaskRunAddedFunc != nil {
					wbr, err := br.onTaskRunAddedFunc(watchTaskRun)
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
				if br.onTaskRunModifiedFunc != nil {
					wbr, err := br.onTaskRunModifiedFunc(watchTaskRun)
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
				if br.onTaskRunDeleteFunc != nil {
					wbr, err := br.onTaskRunDeleteFunc(watchTaskRun)
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

func (tr *TaskRunWatcher) Stop() {
	close(tr.stopCh)
}

func newTaskRunWatcher(
	ctx context.Context,
	trr *resource.Resource,
	listOpts *metav1.ListOptions,
	ioStreams *genericclioptions.IOStreams) (*TaskRunWatcher, error) {
	trw := &TaskRunWatcher{
		ctx:       ctx,
		stopCh:    make(chan bool),
		ioStreams: ioStreams,
	}
	var err error
	trw.watcher, err = trr.Watch(ctx, *listOpts)
	if err != nil {
		return nil, err
	}
	return trw, nil
}
