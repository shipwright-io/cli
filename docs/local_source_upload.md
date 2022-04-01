# Local Source Upload

The purpose of project Shipwright is to transform source code into container images, using known strategies to accomplish this task.

Now, with local source upload, we move Shipwright closer to the developer's inner loop. As a developer, you can use the local source upload feature to stream local content to a Build Controller running in a Kubernetes cluster and create a container image from it. This way, you can try out Shipwright before submitting a pull-request and use the cluster's computing power to build the image.

## Usage

To build an image by using the local source upload feature:

1. Create a Build or use a pre-existing one. Register the standard settings, such as the Build Strategy, on the Build resource. For example:

   ```bash
   shp build create sample-nodejs \
       --source-url="https://github.com/shipwright-io/sample-nodejs.git" \
       --output-image="docker.io/<namespace>/sample-nodejs:latest"
   ```

1. Clone the repository or use a pre-existing one. For example:

   ```bash
   git clone https://github.com/shipwright-io/sample-nodejs.git && \
       cd sample-nodejs
   ```

1. Keep working on the project. When you're ready to build or rebuild a container image with the local changes, run:

   ```bash
   shp build upload sample-nodejs --follow \
       --output-image="docker.io/<namespace>/sample-nodejs:<tag>"
   ```

Notes:

- In the preceding examples, replace placeholders like `<namespace>` and `<tag>` with proper values.
- For `--output-image`, you can specify any Container Registry, use it in combination with `--output-credentials-secret` when needed.

## Streaming

The subcommand `build upload` creates a new `BuildRun` for the informed `Build`. The newly created `BuildRun` contains settings to instruct the Build Controller to wait for the local user upload instead of cloning the external repository as usual.

The command-line interface orchestrates the process of making the `BuildRun`'s Pod wait, and streaming the specified directory when the Pod is ready for it.

The data streamed to the cluster skips the `.git` directory, if present, and any entries specified by the `.gitignore` file.

## Bundling

Alternatively, the `build upload` command can also make use of the `bundle` feature of the Shipwright Build Controller. Instead of a stream into the build pod, with bundle images the local source code is packed (bundled) together into a container image and then pushed into a container registry. The Pod created as a result of the `BuildRun` will pull this image and extract its content. Please note, if the container registry being used is a separate service, make sure to use private images and authentication to protect the source code.

The bundling feature is used in case the `Build` configures a source bundle image name in the source section. This also needs to have a reference to a secret with the credentials for private images.

Files that should be ignored by the bundle feature need to be configured in `.shpignore` file. It follows Git ignore [patterns](https://git-scm.com/docs/gitignore#_pattern_format).
