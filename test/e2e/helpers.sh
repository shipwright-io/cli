#!/usr/bin/env bash

set -eu

BIN="${BIN:-./_output/shp}"

function fail () {
	echo $* >&2
	exit 1
}

function shp () {
	if [ ! -x "${BIN}" ] ; then
		fail "Unable to find '${BIN}' executable"
	fi

	${BIN} ${*}
}

# generate a random string of no more than 16 characters
function random_name () {
	LC_ALL=C tr -dc a-z </dev/urandom | head -c16
}
