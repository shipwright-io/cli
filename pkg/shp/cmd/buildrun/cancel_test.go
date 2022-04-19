package buildrun

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/spf13/cobra"

	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/build/pkg/client/clientset/versioned/fake"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

func TestCancelBuildRun(t *testing.T) {
	tests := map[string]struct {
		br              *v1alpha1.BuildRun
		expectCancelSet bool
		expectErr       bool
	}{
		"does-not-exist": {
			br:        nil,
			expectErr: true,
		},
		"completed": {
			br: &v1alpha1.BuildRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "completed",
					Namespace: metav1.NamespaceDefault,
				},
				Status: v1alpha1.BuildRunStatus{
					Conditions: v1alpha1.Conditions{
						{
							Type:   v1alpha1.Succeeded,
							Status: corev1.ConditionTrue,
						},
					},
				}},
			expectErr: true,
		},
		"failed": {
			br: &v1alpha1.BuildRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "failed",
					Namespace: metav1.NamespaceDefault,
				},
				Status: v1alpha1.BuildRunStatus{
					Conditions: v1alpha1.Conditions{
						{
							Type:   v1alpha1.Succeeded,
							Status: corev1.ConditionFalse,
						},
					},
				}},
			expectErr: true,
		},
		"condition-missing": {
			br: &v1alpha1.BuildRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "failed",
					Namespace: metav1.NamespaceDefault,
				},
			},
			expectCancelSet: true,
		},
		"in-progress": {
			br: &v1alpha1.BuildRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "failed",
					Namespace: metav1.NamespaceDefault,
				},
				Status: v1alpha1.BuildRunStatus{
					Conditions: v1alpha1.Conditions{
						{
							Type:   v1alpha1.Succeeded,
							Status: corev1.ConditionUnknown,
						},
					},
				}},
			expectCancelSet: true,
		},
	}
	for testName, test := range tests {
		t.Logf("running %s with args %#v", testName, test)

		cmd := CancelCommand{cmd: &cobra.Command{}}
		var clientset *fake.Clientset
		if test.br != nil {
			cmd.name = test.br.Name
			clientset = fake.NewSimpleClientset(test.br)
		} else {
			clientset = fake.NewSimpleClientset()
		}

		// set up context
		cmd.Cmd().ExecuteC()
		param := params.NewParamsForTest(nil, clientset, nil, metav1.NamespaceDefault, nil, nil)

		ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
		err := cmd.Run(param, &ioStreams)

		if err != nil && !test.expectErr {
			t.Errorf("%s: did not expect err: %s", testName, err.Error())
		}
		if err == nil && test.expectErr {
			t.Errorf("%s: did not get err when expected", testName)
		}
		if err != nil && test.expectErr {
			continue
		}

		buildRun, _ := clientset.ShipwrightV1alpha1().BuildRuns(param.Namespace()).Get(context.Background(), test.br.Name, metav1.GetOptions{})

		if test.expectCancelSet && !buildRun.IsCanceled() {
			t.Errorf("%s: cancel not set", testName)
		}

		if !test.expectCancelSet && buildRun.IsCanceled() {
			t.Errorf("%s: cancel set", testName)
		}
	}
}
