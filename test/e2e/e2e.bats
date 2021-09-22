#!/usr/bin/env bats

source test/e2e/helpers.sh

@test "shp binary can be executed" {
	result="$(shp --help)"
	[ ! -z "$result" ]
}
