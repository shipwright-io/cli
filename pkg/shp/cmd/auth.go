package cmd

import (
	// IBM Cloud Kubernetes service requirement
	// https://github.com/kubernetes/client-go/issues/345
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	// Google Cloud Platform Kubernetes service requirement
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)
