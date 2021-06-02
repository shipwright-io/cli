package build

import (
	tkncli "github.com/tektoncd/cli/pkg/cli"
	tknlog "github.com/tektoncd/cli/pkg/log"
	tknopt "github.com/tektoncd/cli/pkg/options"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Tail(logOpts *tknopt.LogOptions) error {
	lr, err := tknlog.NewReader(tknlog.LogTypeTask, logOpts)
	if err != nil {
		return err
	}
	logC, errC, err := lr.Read()
	if err != nil {
		return err
	}
	tknlog.NewWriter(tknlog.LogTypeTask).Write(logOpts.Stream, logC, errC)

	return nil
}

func getTKNLogOpts(params tkncli.Params, ioStreams *genericclioptions.IOStreams, taskRunName string) *tknopt.LogOptions {
	return &tknopt.LogOptions{
		AllSteps:    false,
		Follow:      true,
		Params:      params,
		TaskrunName: taskRunName,
		Stream: &tkncli.Stream{
			In:  ioStreams.In,
			Out: ioStreams.Out,
			Err: ioStreams.ErrOut,
		},
	}

}
