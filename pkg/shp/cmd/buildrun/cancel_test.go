package buildrun

import (
	"context"
	"os"
	"testing"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	testflags "github.com/shipwright-io/cli/test/flags"
)

func Test_BuildRunCancelRequiredFlags(t *testing.T) {

	tests := []struct {
		name        string
		args        []string
		completeErr string
		executeErr  string
	}{
		{
			name:        "required flags no name",
			args:        []string{},
			completeErr: `argument list is empty`,
			executeErr:  `accepts 1 arg(s), received 0`,
		},
	}
	for _, tt := range tests {
		o := &BuildRunCancelOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildRunCancelCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Errorf("unexpected error occurred parsing flags: %#v", err)
			}

			err = o.Complete(tt.args)
			if result := testflags.CheckError(err, "Complete", tt.completeErr); len(result) != 0 {
				t.Error(result)
			}

			cmd.SetArgs(tt.args)
			_, err = cmd.ExecuteC()
			if result := testflags.CheckError(err, "Execute", tt.executeErr); len(result) != 0 {
				t.Error(result)
			}

		})
	}
}

// TODO: Fix broken test
/*func TestCancelBuildRun(t *testing.T) {
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

		kclientset := kfake.NewSimpleClientset()

		var clientset *fake.Clientset
		if test.br != nil {
			clientset = fake.NewSimpleClientset(test.br)
		} else {
			clientset = fake.NewSimpleClientset()
		}

		p := params.NewParamsForTest(kclientset, clientset, nil, metav1.NamespaceDefault)

		ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()

		cmd := NewBuildRunCancelCmd(context.Background(), &ioStreams, p)
		if test.br != nil {
			cmd.SetArgs([]string{test.br.Name})
		}
		_, err := cmd.ExecuteC()

		if err != nil && !test.expectErr {
			t.Errorf("%s: did not expect err: %s", testName, err.Error())
		}
		if err == nil && test.expectErr {
			t.Errorf("%s: did not get err when expected", testName)
		}
		if err != nil && test.expectErr {
			continue
		}

		buildRun, _ := clientset.ShipwrightV1alpha1().BuildRuns(p.Namespace()).Get(context.Background(), test.br.Name, metav1.GetOptions{})

		if test.expectCancelSet && !buildRun.IsCanceled() {
			t.Errorf("%s: cancel not set", testName)
		}

		if !test.expectCancelSet && buildRun.IsCanceled() {
			t.Errorf("%s: cancel set", testName)
		}
	}
}*/
