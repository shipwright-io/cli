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

@test "shp environment variable lifecycle" {
	# generate random names for our build and buildrun
	build_name=$(random_name)
	buildrun_name=$(random_name)

	# create a Build with two environment variables
	run shp build create ${build_name} --source-git-url=https://github.com/shipwright-io/sample-go --output-image=my-image --env=VAR_1=build-value-1 --env=VAR_2=build-value-2
	assert_success

	# ensure that the build was successfully created
	assert_output --partial "Created build \"${build_name}\""

	# get the yaml for the Build object
	run kubectl get builds.shipwright.io/${build_name} -o yaml
	assert_success

	# ensure that the environment variables were inserted into the Build object
	assert_output --partial "VAR_1" && assert_output --partial "build-value-1"
	assert_output --partial "VAR_2" && assert_output --partial "build-value-2"

	# create a BuildRun with two environment variables
	run shp buildrun create ${buildrun_name} --buildref-name=${build_name} --env=VAR_2=buildrun-value-2 --env=VAR_3=buildrun-value-3
	assert_success

	# ensure that the build was successfully created
	assert_output --partial "BuildRun created \"${buildrun_name}\" for Build \"${build_name}\""

	# get the yaml for the Build object
	run kubectl get buildruns.shipwright.io/${buildrun_name} -o yaml
	assert_success


	# ensure that the environment variables were inserted into the Build object
	assert_output --partial "VAR_2" && assert_output --partial "buildrun-value-2"
	assert_output --partial "VAR_3" && assert_output --partial "buildrun-value-3"

	# get the taskrun that we created
	run kubectl get taskruns.tekton.dev --selector=buildrun.shipwright.io/name=${buildrun_name} -o name
	assert_success

	run kubectl get ${output} -o yaml
	assert_success

	# ensure that the environment variables were inserted into the TaskRun from the Build object
	assert_output --partial "VAR_1" && assert_output --partial "build-value-1"
	# ensure that the environment variables where inserted into the TaskRun from the BuildRun Object
	# and that the BuildRun environment variable overwrote the Build environment variable
	assert_output --partial "VAR_2" && assert_output --partial "buildrun-value-2"
	assert_output --partial "VAR_3" && assert_output --partial "buildrun-value-3"
}