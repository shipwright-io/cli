package buildrun

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
	shputil "github.com/shipwright-io/cli/pkg/shp/util"
)

// GatherCommand struct stores user input for gather subcommand.
type GatherCommand struct {
	cmd *cobra.Command

	name string
	outputDir string
	archive bool
}
var taskRunGVR = schema.GroupVersionResource{
	Group:    "tekton.dev",
	Version:  "v1",
	Resource: "taskruns",
}

const gatherLongDesc = `
Gather collects the BuildRun object, the TaskRun created for it, the Pod created
for that TaskRun, and all the container logs into a single directory.

By default the command writes:

  buildrun.yaml
  taskrun.yaml
  pod.yaml
  logs/*.log

Use --archive to package the gathered files as a .tar.gz archive.
`

func gatherCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use: "gather <name>",
		Short: "Gather BuildRun diagnostics into a single directory or archive.",
		Long: gatherLongDesc,
		Args: cobra.ExactArgs(1),
	}

	gatherCommand := &GatherCommand{
		cmd: cmd,
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
		return fmt.Errorf("Buildrun name is required")
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

	var taskRunObj *unstructured.Unstructured
	executorKind, executorName := executorForBuildRun(buildRun)
	podName := ""

	switch executorKind {
	case "":
   		fmt.Fprintf(ioStreams.ErrOut, "warning: BuildRun %q does not reference an executor yet\n", c.name)
	case "TaskRun":
		dynamicClient, err := p.DynamicClientSet()
		if err != nil {
			return err
		}

		taskRunObj, err = dynamicClient.Resource(taskRunGVR).Namespace(namespace).Get(ctx, executorName, metav1.GetOptions{})
		switch{
		case err == nil:
			if err = writeYAMLFile(filepath.Join(targetDir, "taskrun.yaml"), taskRunObj.Object); err != nil {
				return err
			}
			if name, found, nestedErr := unstructured.NestedString(taskRunObj.Object, "status", "podName"); nestedErr == nil && found {
				podName = name
			}
		case k8serrors.IsNotFound(err):
			fmt.Fprintf(ioStreams.ErrOut, "warning: TaskRun %q referenced by BuildRun %q was not found\n", executorName, c.name)
		default:
			return err			
		}
	case "PipelineRun":
		return fmt.Errorf("BuildRun %q uses executor kind %q, which gather does not support yet", c.name, executorKind)
	default:
	    return fmt.Errorf("BuildRun %q uses unsupported executor kind %q", c.name, executorKind)
	}


	pod , err := resolvePod(ctx, kubeClient, namespace, c.name, podName) 
	if err != nil {
		return err
	}

	if pod == nil {
		fmt.Fprintf(ioStreams.ErrOut, "warning: no Pod found for BuildRun %q\n", c.name)
	} else {
		if err := writeYAMLFile(filepath.Join(targetDir, "pod.yaml"), pod); err != nil {
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
			if err := os.WriteFile(logPath, []byte(logText), 0o644); err != nil {
				return fmt.Errorf("Failed to write logs %w", err)
			}
		}
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

func resolvePod(ctx context.Context, client kubernetes.Interface, namespace string, buildRunName string, podName string) (*corev1.Pod, error){
	if podName != "" {
		pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		switch {
		case err == nil:
				return pod, nil
		case k8serrors.IsNotFound(err):
		default:
			return nil, err
		}
	}

	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", buildv1beta1.LabelBuildRun, buildRunName),
	})
	if err != nil {
		return nil, err
	}

	if len(podList.Items) == 0 {
		return nil, nil
	}

	return &podList.Items[0], nil
}


func executorForBuildRun(br *buildv1beta1.BuildRun)  (kind string, name string) {
	if br == nil {
		return "", ""
	}

	if br.Status.Executor != nil && br.Status.Executor.Name != ""{
		return br.Status.Executor.Kind, br.Status.Executor.Name
	}



	// backward compatibility with older BuildRuns
	if br.Status.TaskRunName != nil && *br.Status.TaskRunName != "" {
		return "TaskRun", *br.Status.TaskRunName
	}

	return "", ""
}

func createOutputDir(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("filepath %s already exists", path)
	}else if !os.IsNotExist(err) {
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

func createTargz(sourceDir, archivePath string) error {
	// #nosec G304 -- create the file in path provided by the user
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Tar handles structure by filepath so skip directories
		if info.IsDir() {
			return nil
		}

		relpath, err := filepath.Rel(sourceDir, path)
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

		// #nosec G304 
		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}

		defer sourceFile.Close()

		_, err = io.Copy(tarWriter, sourceFile)
		return err
	})
}
