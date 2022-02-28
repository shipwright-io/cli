## shp buildrun

Manage BuildRuns

```
shp buildrun [flags]
```

### Options

```
  -h, --help   help for buildrun
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

* [shp](shp.md)	 - Command-line client for Shipwright's Build API.
* [shp buildrun cancel](shp_buildrun_cancel.md)	 - Cancel BuildRun
* [shp buildrun create](shp_buildrun_create.md)	 - Creates a BuildRun instance.
* [shp buildrun delete](shp_buildrun_delete.md)	 - Delete BuildRun
* [shp buildrun list](shp_buildrun_list.md)	 - List Builds
* [shp buildrun logs](shp_buildrun_logs.md)	 - See BuildRun log output

