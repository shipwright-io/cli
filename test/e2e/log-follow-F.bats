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

@test "shp buildrun logs follow verification" {
  	# generate random names for our build and buildrun
  	build_name=$(random_name)
  	buildrun_name=$(random_name)
    output_image=$(get_output_image build-e2e)

    # create a Build with two environment variables
    run shp build create ${build_name} \
        --source-url=https://github.com/shipwright-io/sample-go \
        --source-context-dir=source-build \
        --output-image=${output_image}
    assert_success

    # initiate a BuildRun
    run shp buildrun create --buildref-name ${build_name} ${buildrun_name}
    # tail logs with -F
    run shp buildrun logs -F ${buildrun_name}
    assert_success


    # confirm output that would only exist if following BuildRun logs
    assert_output --partial "[source-default]"
    assert_output --partial "[build-and-push]"
    assert_output --partial "has succeeded!"
}