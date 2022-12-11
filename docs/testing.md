# Testing

All testing tooling is managed by the `Makefile` on this project, therefore all you'll have to do is
invoke one of those targets to test current changes.

## Unit

To execute unit-testing present on this project, run:

```sh
make test-unit
```

## E2E

End-to-End (E2E) tests aim to mimic the real world usage of this CLI against the
[Shipwright Build Controller][shipwrightBuild]. To run end-to-end tests, make sure you have the
latest changes compiled (`make build`), and then run the tests (`make test-e2e`), in short:

```sh
make build test-e2e
```

### Requirements

The CLI will interact with [Shipwright Build][shipwrightBuild] in the Kubernetes instance. Before running the end-to-end tests make sure [Shipwright Build][shipwrightBuild] and its dependencies is installed.

### GitHub Action

Continuous integration (CI) tests are managed by GitHub Actions, to run these jobs locally please consider [this documentation section][shpSetupContributing] which describes the dependencies and settings required.

After you're all set, run the following target to execute all jobs:

```bash
make act
```

 The CI jobs will exercise [unit](#unit) and [end-to-end](#e2e) tests sequentially, alternatively you may want to run a single suite at the time, that can be achieved using `ARGS` (on `act` target).

 Using `ARGS` you can run a single `job` passing flags to [act][nektosAct] command-line. For instance, in order to only run [unit tests](#unit) (`make test-unit`):

```bash
make act ARGS="--job=test-unit"
```

#### BATS

[BATS][batsCore] is a testing framework for Bash. It's structured as a regular script with enhanced
syntax to define test cases, collect results, and more.

The BATS core tool and it's supporting libraries are linked in this project as `git submodules`,
alleviating the need to install BATS yourself.

To run BATS based tests, make sure that you have cloned the project (including all submodules):

```sh
# Option 1: Clone the repository and populate submodules right away
git clone --recurse-submodules --depth 1 https://github.com/shipwright-io/cli.git

# Option 2: Populate submodules after cloning
cd /your/project/directory/cli
git submodule init
git submodule update
```

then run:

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
[shpSetupContributing]: https://github.com/shipwright-io/setup/blob/main/README.md#contributing
[nektosAct]: https://github.com/nektos/act