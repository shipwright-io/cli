package bundle

import (
	"context"
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	progressbar "github.com/schollz/progressbar/v3"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	buildbundle "github.com/shipwright-io/build/pkg/bundle"
	buildclientset "github.com/shipwright-io/build/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetSourceBundleImage returns the source bundle image of the build that is
// associated with the provided buildrun, an empty string if source bundle is
// not used, or an error in case the build cannot be obtained
func GetSourceBundleImage(ctx context.Context, client buildclientset.Interface, buildRun *buildv1alpha1.BuildRun) (string, error) {
	if buildRun == nil {
		return "", fmt.Errorf("no buildrun provided, given reference is nil")
	}

	if buildRun.Spec.BuildRef != nil {
		name, namespace := buildRun.Spec.BuildRef.Name, buildRun.Namespace

		build, err := client.ShipwrightV1alpha1().Builds(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		if build.Spec.Source.BundleContainer != nil && build.Spec.Source.BundleContainer.Image != "" {
			return build.Spec.Source.BundleContainer.Image, nil
		}
	}

	return "", nil
}

// Push bundles the provided local directory into a container image and pushes
// it to the given registry. For this to work, it relies on valid and working
// container registry access credentials and tokens to be available in the
// local system, for example logins done by `docker login` or similar.
func Push(ctx context.Context, io *genericclioptions.IOStreams, localDirectory string, targetImage string) (name.Digest, error) {
	tag, err := name.NewTag(targetImage)
	if err != nil {
		return name.Digest{}, err
	}

	// The default keychain resolver takes the provided image reference and
	// checks it against the available login credentials in the system. The
	// needs to have done a `docker login` or similar to the respective
	// registry before using the Shipwright CLI.
	auth, err := authn.DefaultKeychain.Resolve(tag.Context())
	if err != nil {
		return name.Digest{}, err
	}

	updates := make(chan v1.Update, 1)
	done := make(chan struct{}, 1)
	go func() {
		var progress *progressbar.ProgressBar
		for {
			select {
			case <-ctx.Done():
				return

			case <-done:
				return

			case update, ok := <-updates:
				if !ok {
					return
				}

				if progress == nil {
					progress = progressbar.NewOptions(int(update.Total),
						progressbar.OptionSetWriter(io.ErrOut),
						progressbar.OptionEnableColorCodes(true),
						progressbar.OptionShowBytes(true),
						progressbar.OptionSetWidth(15),
						progressbar.OptionSetPredictTime(false),
						progressbar.OptionSetDescription("Uploading local source..."),
						progressbar.OptionSetTheme(progressbar.Theme{
							Saucer:        "[green]=[reset]",
							SaucerHead:    "[green]>[reset]",
							SaucerPadding: " ",
							BarStart:      "[",
							BarEnd:        "]"}),
						progressbar.OptionOnCompletion(func() {
							fmt.Fprintln(io.Out)
						}),
					)
					defer progress.Close()
				}

				progress.ChangeMax64(update.Total)
				_ = progress.Set64(update.Complete)
			}
		}
	}()

	fmt.Fprintf(io.Out, "Bundling %q as %q ...\n", localDirectory, targetImage)
	digest, err := buildbundle.PackAndPush(
		tag,
		localDirectory,
		remote.WithContext(ctx),
		remote.WithAuth(auth),
		remote.WithProgress(updates),
	)

	done <- struct{}{}
	return digest, err
}
