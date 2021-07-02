<p align="center">
    <a alt="GitHub-Actions unit-tests" href="https://github.com/shipwright-io/cli/actions">
        <img src="https://github.com/shipwright-io/cli/actions/workflows/unit.yaml/badge.svg">
    </a>
    <a alt="go.pkg.dev project documentation" href="https://pkg.go.dev/mod/github.com/shipwright-io/cli">
        <img src="https://img.shields.io/badge/go.pkg.dev-docs-007d9c?logo=go&logoColor=white">
    </a>
    <a alt="goreportcard.com project report" href="https://goreportcard.com/report/github.com/shipwright-io/cli">
        <img src="https://goreportcard.com/badge/github.com/shipwright-io/cli">
    </a>
</p>

`shp`
-----------

This is an implementation of a command-line client for
[Shipwright's Build](shipwrightbuild) operator, which uses the same pattern as `kubectl`/`oc` and
can be used as a standalone binary and as a `kubectl` plugin.

You can expect to read developer's documentation below, the focus on final users with extend
usage documentation will come in the near future.

## Build and Install

To install it run:

```sh
go get -u github.com/shipwright-io/cli/cmd/shp
```

Or clone the repository, and run `make` to build the binary at `_output` directory:

```sh
make
```

Installation under `/usr/local/bin` is executed via `make install` target.

#### As a `kubectl` Plugin

In order to compile the project as a [kubectl plugin][kubectlplugin], run:

```sh
make kubectl-install
```

By having a `kubectl` named binary in `$PATH`, it behaves a plugin. So, run `kubectl shp` in your
terminal afterwards.


### Run

You can also use `make run`, to `go run` the command-line project directly. For instance:

```sh
make run ARGS='--help'
make run ARGS='run build --help'
```

### Testing

To execute all tests available in the project, run:

```sh
make test
```

## Usage

TBD.

## Project Structure

The project is divided in `cmd` and `pkg` sub-folders, where the `cmd` is a straight forward `main`
for the command-line. The primary business logic lives under `pkg` directory.

The most relevant packages are:

- **`pkg/shp/cmd`**: contains the root `cobra.Command`, and wires up all other CLI sub-commands. It
describes the `SubCommand` interface, and contains a `Runner` implementing sub-commands lifecycle;
- **`pkg/shp/buildrun`**: contains the actions taken against `BuildRun` resources, therefore commands
like `run build` and `create build-run` are handled by this package, which implements `SubCommand`
interface;
- **`pkg/shp/flags`**: contains a Golang flag generator for `BuildRun` resources, and will store all
other flags that could be reused in more than one package;
- **`pkg/shp/util`**: contains utility functions and definitions used in most packages;

You can also read the full Golang docs [here][gopkgdev].

### Upstream Packages

The CLI follows the same structure than `kubectl`/`oc`, and also uses a number of the base packages
employed on those projects.

To name a few of them:

- [`k8s.io/cli-runtime/pkg/genericclioptions`][genericclioptions]: defines the global flags needed
to stablish connection with the API-Server;
- [`k8s.io/kubectl/pkg/cmd/util`][kubectlutil]: exposes a factory interface with helpers to use
global flags to create the Kubernetes client configuration, and instantiate the API-Server client
itself;
- [`k8s.io/kubectl/pkg/util/templates`][kubectltmpl]: helps to organize the command-line interface
in categories;

### Sub-Commands

In addition to upstream `kubectl` packages, `shp` also implements the same lifecycle for
sub-commands, implemented on this project as `SubCommand` interface and `Runner`. The interface
defines how the sub-command package is structured, and `Runner` calls functions in a pre-defined
sequence feeding a instantiated API client and information available during execution.

A more complete example of sub-command implementation can be taken from `pkg/shp/buildrun` package,
where you can observe interactions with the API-Server. Furthermore, package `pkg/shp/initialize` is
placeholder for a future `shp init`.

#### Complete, Validate and Run

The `SubCommand` interface defines the lifecycle of the components, as per:

- `Complete()`: will collect the values in command-line flags, and "complete" the context need for
the upcoming steps. It can also return errors when there's not enough data, or not able to perform
actions;
- `Validate()`: given all necessary data have been collected by previous step (`Complete()`), now is
time to apply validation logic against the context accumulated, and return errors accordingly;
- `Run()`: execute the primary sub-command logic;

<hr/>

## What's Next?

This project will be evaluated by the Shipwright contributors. If we agree on this proof-of-concept
as a starting point, the code base will be moved to a new organization and all current commits should
be rewritten during the migration process.

[shipwrightbuild]: https://github.com/shipwright-io/build/
[kubectlplugin]: https://krew.sigs.k8s.io/docs/developer-guide/
[gopkgdev]: https://pkg.go.dev/mod/github.com/shipwright-io/cli/
[genericclioptions]: https://pkg.go.dev/k8s.io/cli-runtime@v0.17.6/pkg/genericclioptions?tab=overview
[kubectlutil]: https://pkg.go.dev/k8s.io/kubectl@v0.17.6/pkg/cmd/util?tab=overview
[kubectltmpl]: https://pkg.go.dev/k8s.io/kubectl@v0.17.6/pkg/util/templates?tab=overview
