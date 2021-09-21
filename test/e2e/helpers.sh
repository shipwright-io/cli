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