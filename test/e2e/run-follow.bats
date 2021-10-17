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

    # create a Build with two environment variables
    run shp build create ${build_name} --source-url=https://github.com/shipwright-io/sample-go --output-image=my-image
    assert_success

    # initiate a BuildRun with -F
    run shp build run ${build_name} -F
    #assert_success
    #TODO
    # currently do not check success because the create build sample I pulled from either of the existing e2e's do
    # not run to success, both locally and CI; for the one pulled from envvars:
    #     [build-and-push] ERROR: No buildpack groups passed detection.
    #     [build-and-push] ERROR: Please check that you are running against the correct path.
    #     [build-and-push] ERROR: failed to detect: no buildpacks participating
    # ideally, we pull additional items from shipwright-io/build and sort that out, as it uses https://github.com/shipwright-io/sample-go.
    # All that said, the key element for this e2e, log following, is still verifiable


    # confirm output that would only exist if following BuildRun logs
    assert_output --partial "[source-default]"
    assert_output --partial "[place-tools]"
    assert_output --partial "[build-and-push]"

    # initiate a BuildRun with --follow
    run shp build run ${build_name} --follow
    #TODO see above
    #assert_success

     # confirm output that would only exist if following BuildRun logs
     assert_output --partial "[source-default]"
     assert_output --partial "[place-tools]"
     assert_output --partial "[build-and-push]"
}