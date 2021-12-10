#!/usr/bin/env bash
#
# Installs Shipwright Build Controller and Build-Strategies.
#

set -eu

SHIPWRIGHT_HOST="github.com"
SHIPWRIGHT_HOST_PATH="shipwright-io/build/releases/download"
SHIPWRIGHT_VERSION="${SHIPWRIGHT_VERSION:-v0.7.0}"

echo "# Deploying Shipwright Controller '${SHIPWRIGHT_VERSION}'"

kubectl apply -f "https://${SHIPWRIGHT_HOST}/${SHIPWRIGHT_HOST_PATH}/${SHIPWRIGHT_VERSION}/release.yaml"

echo "# Waiting for Build Controller rollout..."

kubectl --namespace="shipwright-build" rollout status deployment shipwright-build-controller --timeout=1m

echo "# Installing upstream Build-Strategies..."

kubectl apply -f "https://${SHIPWRIGHT_HOST}/${SHIPWRIGHT_HOST_PATH}/nightly/default_strategies.yaml"
