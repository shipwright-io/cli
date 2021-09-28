#!/usr/bin/env bats

source test/e2e/helpers.sh

@test "shp binary can be executed" {
	result="$(shp --help)"
	[ ! -z "$result" ]
}

@test "shp build create succesfully run with 1 argument" {
	build_name=$(random_name)
	result="$(shp build create ${build_name} --source-url=url --output-image=image)"
        [ ! -z "$result" ]
	result=$(kubectl get build.shipwright.io ${build_name})
	[ ! -z "$result" ]
}

@test "shp buildrun create succesfully run with 1 argument" {
        build_name=$(random_name)
        result="$(shp build create ${build_name} --source-url=url --output-image=image)"
        [ ! -z "$result" ]
	buildrun_name=$(random_name)
        result="$(shp buildrun create ${buildrun_name} --buildref-name={$build_name})"
	result=$(kubectl get buildrun.shipwright.io ${buildrun_name})
        [ ! -z "$result" ]
}	
