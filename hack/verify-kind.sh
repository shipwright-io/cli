#!/usr/bin/env bash
#
# Make sure a KinD instance is up and running.
#

set -eu

function node_status () {
	echo $(kubectl get node kind-control-plane -o json | \
		jq -r .'status.conditions[] | select(.type == "Ready") | .status')
}

echo "# Using KinD context..."
kubectl config use-context "kind-kind"

echo "# KinD nodes:"
kubectl get nodes

if [ "$(node_status)" == "True" ]; then
	echo "# Kind is Ready!"
else
	echo "# Node is not ready:"
	kubectl describe node kind-control-plane

	echo "# Pods:"
	kubectl get pod -A
	echo "# Events:"
	kubectl get events -A

	exit 1
fi
