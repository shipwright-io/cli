#!/usr/bin/env bats

source test/e2e/helpers.sh

setup() {
	load 'bats/support/load'
	load 'bats/assert/load'
	load 'bats/file/load'
}

teardown() {
	run kubectl delete builds.shipwright.io --all
	run kubectl delete buildruns.shipwright.io --all
}

@test "shp binary can be executed" {
	run shp --help
	assert_success
}

@test "shp --help lists all available commands" {
	run shp --help
	assert_success
	assert_line "Available Commands:"
	assert_line "  build       Manage Builds"
	assert_line "  buildrun    Manage BuildRuns"
	assert_line "  completion  generate the autocompletion script for the specified shell"
	assert_line "  help        Help about any command"
}

@test "shp --help lists some common flags" {
	run shp --help
	assert_success
	assert_line --regexp "-s, --server string    [ ]+The address and port of the Kubernetes API server"
	assert_line --regexp "--user string          [ ]+The name of the kubeconfig user to use"
	assert_line --regexp "--token string         [ ]+Bearer token for authentication to the API server"
	assert_line --regexp "-n, --namespace string [ ]+If present, the namespace scope for this CLI request"
}

@test "shp --help lists also logging flags" {
	run shp --help
	assert_success
	assert_line --regexp "-v, --v Level     [ ]+number for the log level verbosity"
	assert_line --regexp "--log_file string [ ]+If non-empty, use this log file"
}

@test "shp [build/buildrun] create should not error when a name is specified" {
	# generate random names for our build and buildrun
	build_name=$(random_name)
	buildrun_name=$(random_name)

	# ensure that shp build create does not give an error when a build_name is specified
	run shp build create ${build_name} --source-url=url --output-image=image
	assert_success

	# ensure that shp buildrun create does not give an error when a buildrun_name is specified
	run shp buildrun create ${buildrun_name} --buildref-name=${build_name}
	assert_success
}

@test "shp -v=10 build list can log the kubernetes api communication" {
	# ensure that shp command doesn't log the api calls by default
	run shp build list
	assert_success
	refute_line --regexp "GET .*/apis/shipwright.io/v1alpha1/namespaces/default/builds"
	refute_line --partial "Response Headers"
	refute_line --partial "Response Body"

	# ensure that shp command supports -v loglevel flag.
	run shp -v=10 build list
	assert_success
	assert_line --regexp "GET .*/apis/shipwright.io/v1alpha1/namespaces/default/builds"
	assert_line --partial "Response Headers"
	assert_line --partial "Response Body"
}
