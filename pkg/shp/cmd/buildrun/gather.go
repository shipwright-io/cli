package buildrun

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
	shputil "github.com/shipwright-io/cli/pkg/shp/util"
)

// GatherCommand struct stores user input for gather subcommand.
type GatherCommand struct {
	cmd *cobra.Command

	name      string
	outputDir string
	archive   bool
}

var taskRunGVR = schema.GroupVersionResource{
	Group:    "tekton.dev",
	Version:  "v1",
	Resource: "taskruns",
}

var pipelineRunGVR = schema.GroupVersionResource{
	Group:    "tekton.dev",
	Version:  "v1",
	Resource: "pipelineruns",
}

const gatherLongDesc = `
Gather collects the BuildRun object, its executor object, related execution
resources, and available container logs into a single directory.

For BuildRuns executed by a TaskRun, the command writes:

  buildrun.yaml
  taskrun.yaml
  pod.yaml
  logs/*.log

For BuildRuns executed by a PipelineRun, the command writes:

  buildrun.yaml
  pipelinerun.yaml
  taskruns/*.yaml
  pods/*.yaml
  logs/<taskrun-name>/*.log

Use --archive to package the gathered files as a .tar.gz archive.
`

func gatherCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "gather <name>",
		Short: "Gather BuildRun diagnostics into a single directory or archive.",
		Long:  gatherLongDesc,
		Args:  cobra.ExactArgs(1),
	}

	gatherCommand := &GatherCommand{
		cmd:       cmd,
		outputDir: ".",
	}
	// archive is by default set to false.
	cmd.Flags().BoolVarP(&gatherCommand.archive, "archive", "z", gatherCommand.archive, "package gathered diagnostics as a .tar.gz archive")
	cmd.Flags().StringVarP(&gatherCommand.outputDir, "output", "o", gatherCommand.outputDir, "directory to write gathered files")

	return gatherCommand
}

// Cmd returns cobra command object
func (c *GatherCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills in data provided by user
func (c *GatherCommand) Complete(_ *params.Params, _ *genericclioptions.IOStreams, args []string) error {
	c.name = args[0]
	return nil
}

// Validate validates data input by user
func (c *GatherCommand) Validate() error {
	if c.name == "" {
		return fmt.Errorf("buildrun name is required")
	}
	if errs := validation.IsDNS1123Subdomain(c.name); len(errs) > 0 {
		return fmt.Errorf("invalid buildrun name %q: %v", c.name, errs)
	}
	if c.outputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}
	return nil
}

// Run executes gather sub-command logic
func (c *GatherCommand) Run(p *params.Params, ioStreams *genericclioptions.IOStreams) error {
	ctx := c.cmd.Context()
	namespace := p.Namespace()

	shpClient, err := p.ShipwrightClientSet()
	if err != nil {
		return err
	}

	kubeClient, err := p.ClientSet()
	if err != nil {
		return err
	}

	buildRun, err := shpClient.ShipwrightV1beta1().BuildRuns(namespace).Get(ctx, c.name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get BuildRun %q: %w", c.name, err)
	}

	targetDir := filepath.Join(c.outputDir, fmt.Sprintf("buildrun-%s-gather", c.name))
	if err := createOutputDir(targetDir); err != nil {
		return err
	}

	logsDir := filepath.Join(targetDir, "logs")
	if err := os.MkdirAll(logsDir, 0o750); err != nil {
		return fmt.Errorf("error in creating logs directory: %w", err)
	}

	if err := writeYAMLFile(filepath.Join(targetDir, "buildrun.yaml"), buildRun); err != nil {
		return err
	}

	executorKind, executorName := executorForBuildRun(buildRun)

	dynamicClient, err := p.DynamicClientSet()
	if err != nil {
		return err
	}

	switch executorKind {
	case "":
		fmt.Fprintf(ioStreams.ErrOut, "warning: BuildRun %q does not reference an executor yet\n", c.name)
	case "TaskRun":
		err := c.gatherTaskRunExecutor(ctx, dynamicClient, ioStreams, namespace, targetDir, executorName, kubeClient)
		if err != nil {
			return err
		}
	case "PipelineRun":
		err := c.gatherPipelineRunExecutor(ctx, dynamicClient, ioStreams, namespace, targetDir, executorName, kubeClient)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("BuildRun %q uses unsupported executor kind %q", c.name, executorKind)
	}

	finalPath := targetDir
	if c.archive {
		archivePath := finalPath + ".tar.gz"
		if err := createTargz(targetDir, archivePath); err != nil {
			return err
		}

		if err := os.RemoveAll(targetDir); err != nil {
			return err
		}

		finalPath = archivePath
	}

	fmt.Fprintf(ioStreams.Out, "BuildRun diagnostics written to %q\n", finalPath)
	return nil
}

func (c *GatherCommand) gatherTaskRunExecutor(
	ctx context.Context,
	dynamicClient dynamic.Interface,
	ioStreams *genericclioptions.IOStreams,
	namespace string,
	targetDir string,
	executorName string,
	kubeClient kubernetes.Interface,
) error {

	taskRunObj, err := dynamicClient.Resource(taskRunGVR).Namespace(namespace).Get(ctx, executorName, metav1.GetOptions{})
	var podName string
	logsDir := filepath.Join(targetDir, "logs")

	switch {
	case err == nil:
		if err = writeYAMLFile(filepath.Join(targetDir, "taskrun.yaml"), taskRunObj.Object); err != nil {
			return err
		}
		podName = getPodName(taskRunObj)
	case k8serrors.IsNotFound(err):
		fmt.Fprintf(ioStreams.ErrOut, "warning: TaskRun %q referenced by BuildRun %q was not found\n", executorName, c.name)
		return nil
	default:
		return err
	}

	pod, err := resolvePodForTaskRun(ctx, kubeClient, namespace, taskRunObj.GetName(), podName)
	if err != nil {
		return err
	}

	if pod == nil {
		fmt.Fprintf(ioStreams.ErrOut, "warning: no Pod found for BuildRun %q\n", c.name)
	} else {
		if err := writeYAMLFile(filepath.Join(targetDir, "pod.yaml"), pod); err != nil {
			return err
		}

		if err := writePodLogs(ctx, kubeClient, ioStreams, pod, logsDir); err != nil {
			return err
		}
	}

	return nil
}

func (c *GatherCommand) gatherPipelineRunExecutor(
	ctx context.Context,
	dynamicClient dynamic.Interface,
	ioStreams *genericclioptions.IOStreams,
	namespace string,
	targetDir string,
	executorName string,
	kubeClient kubernetes.Interface,
) error {

	pipelineRunObj, err := dynamicClient.Resource(pipelineRunGVR).Namespace(namespace).Get(ctx, executorName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get PipelineRun %q: %w", executorName, err)
	}

	if err := writeYAMLFile(filepath.Join(targetDir, "pipelinerun.yaml"), pipelineRunObj.Object); err != nil {
		return err
	}

	taskrunList, err := dynamicClient.Resource(taskRunGVR).Namespace(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("tekton.dev/pipelineRun=%s", executorName),
	})
	if err != nil {
		return fmt.Errorf("failed to list TaskRuns for PipelineRun %q: %w", executorName, err)
	}
	if len(taskrunList.Items) == 0 {
		fmt.Fprintf(ioStreams.ErrOut, "warning: PipelineRun %q did not produce any TaskRuns yet\n", executorName)
		return nil
	}

	taskRunsDir := filepath.Join(targetDir, "taskruns")
	podsDir := filepath.Join(targetDir, "pods")
	logsDir := filepath.Join(targetDir, "logs")

	for _, dir := range []string{taskRunsDir, podsDir, logsDir} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}

	for _, taskRun := range taskrunList.Items {
		taskRunName := taskRun.GetName()

		taskRunPath := filepath.Join(taskRunsDir, fmt.Sprintf("%s.yaml", taskRunName))
		if err := writeYAMLFile(taskRunPath, taskRun.Object); err != nil {
			return err
		}

		podName := getPodName(&taskRun)

		pod, err := resolvePodForTaskRun(ctx, kubeClient, namespace, taskRunName, podName)
		if err != nil {
			return err
		}

		if pod == nil {
			fmt.Fprintf(ioStreams.ErrOut, "warning: no Pod found for TaskRun %q in PipelineRun %q\n", taskRunName, executorName)
			continue
		}

		podPath := filepath.Join(podsDir, fmt.Sprintf("%s.yaml", pod.Name))
		if err := writeYAMLFile(podPath, pod); err != nil {
			return err
		}

		taskRunLogsDir := filepath.Join(logsDir, taskRunName)
		if err := writePodLogs(ctx, kubeClient, ioStreams, pod, taskRunLogsDir); err != nil {
			return err
		}
	}
	return nil
}

func resolvePodForTaskRun(ctx context.Context, client kubernetes.Interface, namespace string, taskRunName string, podName string) (*corev1.Pod, error) {
	if podName != "" {
		pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err == nil {
			return pod, nil
		}
		// Not Found error is intentionally ignored to fallback to another way of finding the pod.
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
	}

	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("tekton.dev/taskRun=%s", taskRunName),
	})
	if err != nil {
		return nil, err
	}
	if len(podList.Items) == 0 {
		return nil, nil
	}

	return &podList.Items[0], nil
}

func executorForBuildRun(br *buildv1beta1.BuildRun) (kind string, name string) {
	if br == nil {
		return "", ""
	}

	if br.Status.Executor != nil && br.Status.Executor.Name != "" {
		return br.Status.Executor.Kind, br.Status.Executor.Name
	}

	return "", ""
}

func createOutputDir(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("filepath %s already exists", path)
	} else if !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(path, 0o750)
}

func writeYAMLFile(path string, object any) error {
	data, err := yaml.Marshal(object)
	if err != nil {
		return fmt.Errorf("failed to marshal object to yaml: %w", err)
	}
	return os.WriteFile(path, data, 0o600)
}

func createTargz(sourceDir, archivePath string) (err error) {
	// #nosec G304 -- create the file in path provided by the user
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}

	// Delete the partially written archive if something goes wrong.
	defer func() {
		if err != nil {
			_ = file.Close()
			_ = os.Remove(archivePath)
		}
	}()

	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)

	rootFS := os.DirFS(sourceDir)

	// use sourceDir as the root to prevent TOCTOU problems
	err = fs.WalkDir(rootFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Tar handles structure by filepath so skip directories
		if d.IsDir() {
			return nil
		}

		relpath := path

		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Avoid cross platform bugs (windows uses \ slash)
		header.Name = filepath.ToSlash(relpath)

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Open file via rootFS (prevents path escape)
		sourceFile, err := rootFS.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(tarWriter, sourceFile)
		if err != nil {
			_ = sourceFile.Close()
			return err
		}

		if err := sourceFile.Close(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	if err = tarWriter.Close(); err != nil {
		return err
	}
	if err = gzipWriter.Close(); err != nil {
		return err
	}
	return file.Close()
}

func writePodLogs(ctx context.Context, kubeClient kubernetes.Interface, ioStreams *genericclioptions.IOStreams, pod *corev1.Pod, logsDir string) error {
	if err := os.MkdirAll(logsDir, 0o750); err != nil {
		return err
	}
	for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
		logText, err := shputil.GetPodLogs(ctx, kubeClient, *pod, container.Name)
		if err != nil {
			fmt.Fprintf(ioStreams.ErrOut, "warning: could not fetch logs for Pod %q container %q: %s\n",
				pod.Name,
				container.Name,
				err.Error(),
			)
			continue
		}

		logPath := filepath.Join(logsDir, fmt.Sprintf("%s.log", container.Name))
		if err := os.WriteFile(logPath, []byte(logText), 0o600); err != nil {
			return fmt.Errorf("failed to write logs: %w", err)
		}
	}
	return nil
}

func getPodName(obj *unstructured.Unstructured) string {
	if obj == nil {
		return ""
	}
	podName, _, _ := unstructured.NestedString(obj.Object, "status", "podName")
	return podName
}
