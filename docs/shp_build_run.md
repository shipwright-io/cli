## shp build run

Start a build specified by 'name'

### Synopsis


Creates a unique BuildRun instance for the given Build, which starts the build
process orchestrated by the Shipwright build controller. For example:

	$ shp build run my-app


```
shp build run <name> [flags]
```

### Options

```
      --buildref-name string                     name of build resource to reference
  -e, --env stringArray                          specify a key-value pair for an environment variable to set for the build container (default [])
  -F, --follow                                   Start a build and watch its log until it completes or fails.
  -h, --help                                     help for run
      --node-selector stringArray                set of key-value pairs that correspond to labels of a node to match (default [])
      --output-image string                      image employed during the building process
      --output-image-annotation stringArray      specify a set of key-value pairs that correspond to annotations to set on the output image (default [])
      --output-image-label stringArray           specify a set of key-value pairs that correspond to labels to set on the output image (default [])
      --output-image-push-secret string          name of the secret with output image push credentials
      --output-insecure                          flag to indicate an insecure container registry
      --param-value stringArray                  set of key-value pairs to pass as parameters to the buildStrategy (default [])
      --retention-ttl-after-failed duration      duration to delete the BuildRun after it failed
      --retention-ttl-after-succeeded duration   duration to delete the BuildRun after it succeeded
      --sa-name string                           Kubernetes service-account name
      --scheduler-name string                    specify the scheduler to be used to dispatch the Pod
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

