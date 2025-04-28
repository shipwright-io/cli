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

@test "shp output image labels and annotation lifecycle" {
	# generate random names for our build and buildrun
	build_name=$(random_name)
	buildrun_name=$(random_name)

	# create a Build with a label and an annotation
	run shp build create ${build_name} --source-git-url=https://github.com/shipwright-io/sample-go --output-image=my-image --output-image-label=foo=bar --output-image-annotation=created-by=shipwright
	assert_success

	# ensure that the build was successfully created
	assert_output --partial "Created build \"${build_name}\""

	# get the yaml for the Build object
	run kubectl get builds.shipwright.io/${build_name} -o yaml
	assert_success

	# ensure that the label and annotation were inserted into the Build object
	assert_output --partial "foo: bar"
	assert_output --partial "created-by: shipwright"

	# create a BuildRun with two environment variables
    run shp buildrun create ${buildrun_name} --buildref-name=${build_name} --output-image=my-image --output-image-label=foo=bar123 --output-image-annotation=owned-by=shipwright
    assert_success

    # ensure that the build was successfully created
    assert_output --partial "BuildRun created \"${buildrun_name}\" for Build \"${build_name}\""

    # get the yaml for the BuildRun object
    run kubectl get buildruns.shipwright.io/${buildrun_name} -o yaml
    assert_success

    # ensure that the label and annotation were inserted into the BuildRun object
    assert_output --partial "foo: bar123"
    assert_output --partial "owned-by: shipwright"

    # get the taskrun that we created
    run kubectl get taskruns.tekton.dev --selector=buildrun.shipwright.io/name=${buildrun_name} -o name
    assert_success

    run kubectl get ${output} -o yaml
    assert_success

	# ensure that the annotation was inserted into the TaskRun from the Build object which is not in BuildRun
	assert_output --partial "created-by=shipwright"
	# ensure that the labels and annotations where inserted into the TaskRun from the BuildRun Object
	# and that the value from BuildRun override the ones defined in Build
	assert_output --partial "owned-by=shipwright"
	assert_output --partial "foo=bar123"
}