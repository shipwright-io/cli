package bundle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/schollz/progressbar/v3"
	buildbundle "github.com/shipwright-io/build/pkg/bundle"
)

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

// Prune removes the image from the container registry
//
// Deleting a tag, or a whole repo is not as straightforward as initially
// planned as DockerHub seems to restrict deleting a single tag for
// standard users. This might be subject to change, but as of September
// 2021 it is limited to the business tier. However, there is an API call
// to delete the whole repository. In case there is only one tag used in
// a repository, the effect is pretty much the same. For convenience, there
// is a provider switch to deal with images on DockerHub differently.
//
// DockerHub images:
// - In case the repository only has one tag, the repository is deleted.
// - If there are multiple tags, the tag to be deleted is overwritten
//   with an empty image (to remove the content, and save quota).
// - Edge case would be no tags in the repository, which is ignored.
//
// Other registries:
// Use standard spec delete API request to delete the provided tag.
//
func Prune(ctx context.Context, io *genericclioptions.IOStreams, image string) error {
	ref, err := name.ParseReference(image)
	if err != nil {
		return err
	}

	auth, err := authn.DefaultKeychain.Resolve(ref.Context())
	if err != nil {
		return err
	}

	switch ref.Context().RegistryStr() {
	case "index.docker.io":
		list, err := remote.List(ref.Context(), remote.WithContext(ctx), remote.WithAuth(auth))
		if err != nil {
			return err
		}

		switch len(list) {
		case 0:
			return nil

		case 1:
			authr, err := auth.Authorization()
			if err != nil {
				return err
			}

			token, err := dockerHubLogin(authr.Username, authr.Password)
			if err != nil {
				return err
			}

			return dockerHubRepoDelete(token, ref)

		default:
			fmt.Fprintf(io.ErrOut, "Removing a specific image tag is not supported on %s, the respective image tag will be overwritten with an empty image.\n", ref.Context().RegistryStr())

			// In case the input argument included a digest, the reference
			// needs to be updated to exclude the digest for the empty image
			// override to succeed.
			switch ref.(type) {
			case name.Digest:
				tmp := strings.SplitN(image, "@", 2)
				ref, err = name.NewTag(tmp[0])
				if err != nil {
					return err
				}
			}

			return remote.Write(ref, empty.Image, remote.WithContext(ctx), remote.WithAuth(auth))
		}

	default:
		return remote.Delete(
			ref,
			remote.WithContext(ctx),
			remote.WithAuth(auth),
		)
	}
}

func dockerHubLogin(username string, password string) (string, error) {
	type LoginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	loginData, err := json.Marshal(LoginData{Username: username, Password: password})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://hub.docker.com/v2/users/login/", bytes.NewReader(loginData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		type LoginToken struct {
			Token string `json:"token"`
		}

		var loginToken LoginToken
		if err := json.Unmarshal(bodyData, &loginToken); err != nil {
			return "", err
		}

		return loginToken.Token, nil

	default:
		return "", fmt.Errorf(string(bodyData))
	}
}

func dockerHubRepoDelete(token string, ref name.Reference) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/", ref.Context().RepositoryStr()), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "JWT "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusAccepted:
		return nil

	default:
		return fmt.Errorf("failed with HTTP status code %d: %s", resp.StatusCode, string(respData))
	}
}
