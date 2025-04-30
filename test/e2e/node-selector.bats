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

@test "shp build create --node-selector single label" {
    # generate random names for our build
	build_name=$(random_name)

    # create a Build with node selector 
    run shp build create ${build_name} --source-git-url=https://github.com/shipwright-io/sample-go --output-image=my-fake-image --node-selector="kubernetes.io/hostname=node-1"
    assert_success

    # ensure that the build was successfully created
	assert_output --partial "Created build \"${build_name}\""

    # get the jsonpath of Build object .spec.nodeSelector
	run kubectl get builds.shipwright.io/${build_name} -ojsonpath="{.spec.nodeSelector}"
	assert_success

    assert_output '{"kubernetes.io/hostname":"node-1"}'
}

@test "shp build create --node-selector multiple labels" {
    # generate random names for our build
	build_name=$(random_name)

    # create a Build with node selector 
    run shp build create ${build_name} --source-git-url=https://github.com/shipwright-io/sample-go --output-image=my-fake-image --node-selector="kubernetes.io/hostname=node-1" --node-selector="kubernetes.io/os=linux" 
    assert_success

    # ensure that the build was successfully created
	assert_output --partial "Created build \"${build_name}\""

    # get the jsonpath of Build object .spec.nodeSelector
	run kubectl get builds.shipwright.io/${build_name} -ojsonpath="{.spec.nodeSelector}"
	assert_success

    assert_output --partial '"kubernetes.io/hostname":"node-1"'
    assert_output --partial '"kubernetes.io/os":"linux"'
}

@test "shp buildrun create --node-selector single label" {
    # generate random names for our buildrun
	buildrun_name=$(random_name)
	build_name=$(random_name)

    # create a Build with node selector 
    run shp buildrun create ${buildrun_name} --buildref-name=${build_name} --node-selector="kubernetes.io/hostname=node-1"
    assert_success

    # ensure that the build was successfully created
	assert_output --partial "BuildRun created \"${buildrun_name}\" for Build \"${build_name}\""

    # get the jsonpath of Build object .spec.nodeSelector
	run kubectl get buildruns.shipwright.io/${buildrun_name} -ojsonpath="{.spec.nodeSelector}"
	assert_success

    assert_output '{"kubernetes.io/hostname":"node-1"}'
}

@test "shp buildrun create --node-selector multiple labels" {
    # generate random names for our buildrun
	buildrun_name=$(random_name)
	build_name=$(random_name)

    # create a Build with node selector 
    run shp buildrun create ${buildrun_name} --buildref-name=${build_name} --node-selector="kubernetes.io/hostname=node-1"  --node-selector="kubernetes.io/os=linux"
    assert_success

    # ensure that the build was successfully created
	assert_output --partial "BuildRun created \"${buildrun_name}\" for Build \"${build_name}\""

    # get the jsonpath of Build object .spec.nodeSelector
	run kubectl get buildruns.shipwright.io/${buildrun_name} -ojsonpath="{.spec.nodeSelector}"
	assert_success

    assert_output --partial '"kubernetes.io/hostname":"node-1"'
    assert_output --partial '"kubernetes.io/os":"linux"'
}


@test "shp build run --node-selector set" {
    # generate random names for our build
	build_name=$(random_name)

    # create a Build with node selector 
    run shp build create ${build_name} --source-git-url=https://github.com/shipwright-io/sample-go --output-image=my-fake-image
    assert_success

    # ensure that the build was successfully created
	assert_output --partial "Created build \"${build_name}\""

    # get the build object
	run kubectl get builds.shipwright.io/${build_name}
	assert_success

    run shp build run ${build_name} --node-selector="kubernetes.io/hostname=node-1"

    # get the jsonpath of Build object .spec.nodeSelector
	run kubectl get buildruns.shipwright.io -ojsonpath='{.items[*].spec.nodeSelector}' 
	assert_success
    assert_output --partial '"kubernetes.io/hostname":"node-1"'
}