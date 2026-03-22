## shp buildrun gather

Gather BuildRun diagnostics into a single directory or archive.

### Synopsis


Gather collects the BuildRun object, the TaskRun created for it, the Pod created
for that TaskRun, and all the container logs into a single directory.

By default the command writes:

  buildrun.yaml
  taskrun.yaml
  pod.yaml
  logs/*.log

Use --archive to package the gathered files as a .tar.gz archive.


```
shp buildrun gather <name> [flags]
```

### Options

```
  -z, --archive         package gathered diagnostics as a .tar.gz archive
  -h, --help            help for gather
  -o, --output string   directory to write gathered files (default ".")
```

### Options inherited from parent commands

```
      --kubeconfig string        Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string         If present, the namespace scope for this CLI request
      --request-timeout string   The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
```

### SEE ALSO

* [shp buildrun](shp_buildrun.md)	 - Manage BuildRuns

