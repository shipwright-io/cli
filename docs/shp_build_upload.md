## shp build upload

Run a Build with local data

### Synopsis


Creates a new BuildRun instance and instructs the Build Controller to wait for the data streamed,
instead of executing "git clone". Therefore, you can employ Shipwright Builds from a local repository
clone.

The upload skips the ".git" directory completely, and it follows the ".gitignore" directives, when
the file is found at the root of the directory uploaded.

	$ shp buildrun upload <build-name>
	$ shp buildrun upload <build-name> /path/to/repository


```
shp build upload <build-name> [path/to/source|.] [flags]
```

### Options

```
      --buildref-apiversion string            API version of build resource to reference
      --buildref-name string                  name of build resource to reference
  -e, --env stringArray                       specify a key-value pair for an environment variable to set for the build container (default [])
  -F, --follow                                Start a build and watch its log until it completes or fails.
  -h, --help                                  help for upload
      --output-credentials-secret string      name of the secret with builder-image pull credentials
      --output-image string                   image employed during the building process
      --output-image-annotation stringArray   specify a set of key-value pairs that correspond to annotations to set on the output image (default [])
      --output-image-label stringArray        specify a set of key-value pairs that correspond to labels to set on the output image (default [])
      --sa-generate                           generate a Kubernetes service-account for the build
      --sa-name string                        Kubernetes service-account name
      --timeout duration                      build process timeout
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default cache directory (default "~/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
```

### SEE ALSO

* [shp build](shp_build.md)	 - Manage Builds

