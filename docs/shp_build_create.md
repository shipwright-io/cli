## shp build create

Create Build

### Synopsis


Creates a new Build instance using the first argument as its name. For example:

	$ shp build create my-app --source-url="..." --output-image="..."


```
shp build create <name> [flags]
```

### Options

```
  -e, --env stringArray                          specify a key-value pair for an environment variable to set for the build container (default [])
  -h, --help                                     help for create
      --node-selector stringArray                set of key-value pairs that correspond to labels of a node to match (default [])
      --output-image string                      image employed during the building process
      --output-image-annotation stringArray      specify a set of key-value pairs that correspond to annotations to set on the output image (default [])
      --output-image-label stringArray           specify a set of key-value pairs that correspond to labels to set on the output image (default [])
      --output-image-push-secret string          name of the secret with output image push credentials
      --output-insecure                          flag to indicate an insecure container registry
      --param-value stringArray                  set of key-value pairs to pass as parameters to the buildStrategy (default [])
      --retention-failed-limit uint              number of failed BuildRuns to be kept (default 65535)
      --retention-succeeded-limit uint           number of succeeded BuildRuns to be kept (default 65535)
      --retention-ttl-after-failed duration      duration to delete a failed BuildRun after completion
      --retention-ttl-after-succeeded duration   duration to delete a succeeded BuildRun after completion
      --source-context-dir string                use a inner directory as context directory
      --source-git-clone-secret string           name of the secret with credentials to access the git source, e.g. git credentials
      --source-git-revision string               git repository source revision
      --source-git-url string                    git repository source URL
      --source-oci-artifact-image string         source OCI artifact image reference, e.g. ghcr.io/shipwright-io/sample-go/source-bundle:latest
      --source-oci-artifact-prune pruneOption    source OCI artifact image prune option, either Never, or AfterPull (default Never)
      --source-oci-artifact-pull-secret string   name of the secret with credentials to access the OCI artifact image, e.g. registry credentials
      --strategy-kind string                     build-strategy kind (default "ClusterBuildStrategy")
      --strategy-name string                     build-strategy name (default "buildpacks-v3")
      --timeout duration                         build process timeout
```

### Options inherited from parent commands

```
      --kubeconfig string        Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string         If present, the namespace scope for this CLI request
      --request-timeout string   The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
```

### SEE ALSO

* [shp build](shp_build.md)	 - Manage Builds

