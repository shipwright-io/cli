#!/bin/bash
#
# Installs KinD (Kubernetes in Docker) via "go get" and configure it as current context.
#

set -eu

if [ ! -f "${GOPATH}/bin/kind" ] ; then
    echo "# Installing KinD..."
    GO111MODULE=off go get sigs.k8s.io/kind
fi

# print kind version
kind --version

# kind cluster name
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-kind}"

# kind cluster version
KIND_CLUSTER_VERSION="${KIND_CLUSTER_VERSION:-v1.17.0}"

echo "# Creating a new Kubernetes cluster..."
kind create cluster --quiet --name="${KIND_CLUSTER_NAME}" --image="kindest/node:${KIND_CLUSTER_VERSION}" --wait=120s

echo "# Using KinD context..."
kubectl config use-context "kind-kind"

echo "# KinD nodes:"
kubectl get nodes
