# Testing

All testing tooling is managed by the `Makefile` on this project, therefore all you'll have to do is
invoke one of those targets to test current changes.

To run all projet tests, execute:

```sh
make test
```

## Unit-Testing

To execute unit-testing present on this project, run:

```sh
make test-unit
```

## End-to-End (E2E)

End-to-End tests aim to mimic the real world usage of this CLI against the
[Shipwright Build Controller][shipwrightBuild]. To run end-to-end tests, make sure you have the
latest changes compiled (`make build`), and then run the tests (`make test-e2e`), in short:

```sh
make build test-e2e
```

### Requirements

The following componets are required to run end-to-end tests.

#### Kubernetes

Before testing a [KinD][kindSig] instance is created, running a predefined version of Kubernetes.
Please consider [GitHub Action files](../.github/workflows/e2e.yaml) for more details. Additionally,
`kubectl` is required to intereact with the Kubernetes Cluster, you need to install it before running
scripts on the [`./hack` directory](../hack).

#### Shipwright Build Controller

The [Shipwright Build Controller][shipwrightBuild] must be up and running as well. To install the
controller and its dependencies, run:

```sh
make install-shipwright
```

The install script waits for the Controller instance to be running.

#### BATS

[BATS][batsCore] is a testing framework for Bash. It's structured as a regular script with enhanced
syntax to define test cases, collect results, and more.

To run BATS based tests, make sure you have `bats` installed, and then execute:

```sh
make test-e2e
```

### Test Cases

The usual [BATS][batsCore] test-cases can be defined as the following example:

```bash
@test "short test description, or test name" {
	# prepare the test-case context running the necessary commands to do so.
	kubectl ...

	# then execute the actual test, usually by running "shp" command with arguments.
	result=$(shp ...)

	# at the end, assert the test result by inspecting the output, or maybe, execute probes to
	# identify if the desired changes have been performed on the cluster, and such.
	[ ! -z "${result}" ]
}
```

Repetitive tasks can be defined as Bash `function`, the actual `shp` command employed during testing
is overwritten to use executable compiled in the project folder.


[kindSig]: https://kind.sigs.k8s.io/
[batsCore]: https://github.com/bats-core/bats-core
[shipwrightBuild]: https://github.com/shipwright-io/build
