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
	run shp build create ${build_name} --source-url=https://github.com/shipwright-io/sample-go --output-image=my-image --output-image-label=foo=bar --output-image-annotation=created-by=shipwright
	assert_success

	# ensure that the build was successfully created
	assert_output --partial "Created build \"${build_name}\""

	# get the yaml for the Build object
	run kubectl get builds.shipwright.io/${build_name} -o yaml
	assert_success

	# ensure that the label and annotation were inserted into the Build object
	assert_output --partial "foo: bar"
	assert_output --partial "created-by: shipwright"
}