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
