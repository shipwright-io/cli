package build

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/follower"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func runCommandLifecycle(
	t *testing.T,
	runCommand *RunCommand,
	fp *params.FakeParams,
	ioStreams *genericclioptions.IOStreams,
	args []string,
) error {
	var err error
	if err = runCommand.Complete(fp, ioStreams, args); err != nil {
		t.Logf("Complete(): err=%q", err)
		return err
	}
	if err = runCommand.Validate(); err != nil {
		t.Logf("Validate(): err=%q", err)
		return err
	}
	if runCommand.Run(fp, ioStreams); err != nil {
		t.Logf("Run(): err=%q", err)
		return err
	}
	return nil
}
func TestRunCommand_Run(t *testing.T) {
	g := gomega.NewWithT(t)

	tests := []struct {
		name      string
		buildName string
		follow    bool
		wantErr   bool
	}{{
		name:      "build name is empty",
		buildName: "",
		follow:    false,
		wantErr:   true,
	}, {
		name:      "not following logs",
		buildName: "build",
		follow:    false,
		wantErr:   false,
	}, {
		name:      "following logs",
		buildName: "build",
		follow:    true,
		wantErr:   false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fp := params.NewFakeParams(20*time.Second, nil, nil)
			pw, err := reactor.NewPodWatcherFromParams(ctx, fp)
			g.Expect(err).To(gomega.BeNil())

			ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()
			podLogsFollower, err := follower.NewPodLogsFollowerFromParams(ctx, fp, pw, &ioStreams)
			g.Expect(err).To(gomega.BeNil())

			runCommand := &RunCommand{
				cmd: runCmd().Cmd(),
				buildRunSpec: &v1alpha1.BuildRunSpec{
					BuildRef: &v1alpha1.BuildRef{
						Name: tt.buildName,
					},
				},
				follow:          tt.follow,
				podLogsFollower: podLogsFollower,
			}

			err = runCommandLifecycle(t, runCommand, fp, &ioStreams, []string{tt.buildName})
			t.Logf("output=%q", out.String())
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			g.Eventually(func() bool {
				buildClientset, _ := fp.ShipwrightClientSet()
				list, err := buildClientset.ShipwrightV1alpha1().
					BuildRuns(metav1.NamespaceDefault).
					List(ctx, metav1.ListOptions{})
				if err != nil {
					return false
				}
				return len(list.Items) == 1
			}).Should(gomega.BeTrue())
		})
	}
}
