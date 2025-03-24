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

@test "shp build run follow verification" {
  	# generate random names for our build and buildrun
  	build_name=$(random_name)
  	buildrun_name=$(random_name)
    output_image=$(get_output_image build-e2e)

    # create a Build
    run shp build create ${build_name} \
        --source-git-url=https://github.com/shipwright-io/sample-go \
        --source-context-dir=source-build \
        --output-image=${output_image}
    assert_success

    # initiate a BuildRun with -F
    run shp build run ${build_name} -F
    assert_success

    # confirm output that would only exist if following BuildRun logs
    assert_output --partial "[source-default]"
    assert_output --partial "[build-and-push]"
    assert_output --partial "has succeeded!"

    # initiate a BuildRun with --follow
    run shp build run ${build_name} --follow
    assert_success

    # confirm output that would only exist if following BuildRun logs
    assert_output --partial "[source-default]"
    assert_output --partial "[build-and-push]"
    assert_output --partial "has succeeded!"
}

@test "shp build run follow verification with failure" {
  	# generate random names for our build and buildrun
  	build_name=$(random_name)
  	buildrun_name=$(random_name)
    output_image=$(get_output_image build-e2e)

    # create a Build which will fail
    run shp build create ${build_name} \
        --source-git-url=https://github.com/shipwright-io/sample-go \
        --source-context-dir=source-build \
        --output-image=${output_image} \
        --strategy-name buildkit
    assert_success

    # initiate a BuildRun with -F
    run shp build run ${build_name} -F
    assert_failure

    # confirm output that would only exist if following BuildRun logs
    assert_output --partial "[source-default]"
    assert_output --partial "[build-and-push]"

    # confirm failure message
    assert_output --partial "has failed at step \"step-build-and-push\" because of DockerfileNotFound: The Dockerfile '/workspace/source/source-build/Dockerfile' does not exist."
}