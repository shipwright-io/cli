#!/usr/bin/env bash

set -eu

# path to the local build of shp command-line
BIN="${BIN:-./_output/shp}"

# registry hostname and namespace, in order to compose image names on the fly
OUTPUT_HOSTNAME="${OUTPUT_HOSTNAME:-registry.registry.svc.cluster.local:32222}"
OUTPUT_NAMESPACE="${OUTPUT_NAMESPACE:-shipwright-io}"

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

# formats the container image name based on the environment variables.
function get_output_image () {
	local name=${1}
	echo "${OUTPUT_HOSTNAME}/${OUTPUT_NAMESPACE}/${name}"
}
