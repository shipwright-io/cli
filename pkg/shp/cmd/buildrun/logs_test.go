package buildrun

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/follower"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func logCommandLifecycle(
	t *testing.T,
	logsCommand *LogsCommand,
	fp *params.FakeParams,
	ioStreams *genericclioptions.IOStreams,
	args []string,
) error {
	var err error
	if err = logsCommand.Complete(fp, ioStreams, args); err != nil {
		t.Logf("Complete(): err=%q", err)
		return err
	}
	if err = logsCommand.Validate(); err != nil {
		t.Logf("Validate(): err=%q", err)
		return err
	}
	if err = logsCommand.Run(fp, ioStreams); err != nil {
		t.Logf("Run(): err=%q", err)
		return err
	}
	return nil
}

func TestLogsCommand_Run(t *testing.T) {
	g := gomega.NewWithT(t)

	buildName := "build"
	buildRunName := "buildrun"

	br := &v1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      buildRunName,
		},
		Spec: v1alpha1.BuildRunSpec{
			BuildRef: &v1alpha1.BuildRef{
				Name: buildName,
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      buildRunName,
			Labels: map[string]string{
				v1alpha1.LabelBuild:    buildName,
				v1alpha1.LabelBuildRun: buildRunName,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "container",
			}},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodSucceeded,
		},
	}

	tests := []struct {
		name    string
		br      *v1alpha1.BuildRun
		pod     *corev1.Pod
		wantErr bool
	}{{
		name:    "buildrun does not exist",
		br:      nil,
		pod:     nil,
		wantErr: true,
	}, {
		name:    "pod does not exist",
		br:      br,
		pod:     nil,
		wantErr: true,
	}, {
		name:    "buildrun and pod exists",
		br:      br,
		pod:     pod,
		wantErr: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coreObjects := []runtime.Object{}
			if tt.pod != nil {
				coreObjects = append(coreObjects, tt.pod)
			}
			shipwrightObjects := []runtime.Object{}
			if tt.br != nil {
				shipwrightObjects = append(shipwrightObjects, tt.br)
			}

			// setting up the test clients using the objects informed, or empty in some cases
			fp := params.NewFakeParams(20*time.Second, coreObjects, shipwrightObjects)

			ctx := context.Background()
			pw, err := reactor.NewPodWatcherFromParams(ctx, fp)
			g.Expect(err).To(gomega.BeNil())

			ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()
			podLogsFollower, err := follower.NewPodLogsFollowerFromParams(ctx, fp, pw, &ioStreams)
			g.Expect(err).To(gomega.BeNil())

			logsCommand := &LogsCommand{
				cmd:             &cobra.Command{},
				podLogsFollower: podLogsFollower,
			}

			// running ExecuteC to instantiate the Cobra Command context, used on the PodWatcher
			_, err = logsCommand.Cmd().ExecuteC()
			g.Expect(err).To(gomega.BeNil())

			// executing the log subcommand lifecycle, which may return error on any of the steps
			err = logCommandLifecycle(t, logsCommand, fp, &ioStreams, []string{buildRunName})
			t.Logf("output=%q", out.String())

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			g.Expect(logsCommand.buildRunName).To(gomega.Equal(buildRunName))
		})
	}
}
