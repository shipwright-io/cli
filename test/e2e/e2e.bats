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
  run kubectl delete buildstrategies.shipwright.io --all
}

@test "shp binary can be executed" {
  run shp --help
  assert_success
}

@test "shp --help lists all available commands" {
  run shp --help
  assert_success
  assert_line --regexp '^Available Commands:$'
  assert_line --regexp '^[[:space:]]+build[[:space:]]+Manage Builds$'
  assert_line --regexp '^[[:space:]]+buildrun[[:space:]]+Manage BuildRuns$'
  assert_line --regexp '^[[:space:]]+buildstrategy[[:space:]]+Manage namespaced BuildStrategies$'
  assert_line --regexp '^[[:space:]]+clusterbuildstrategy[[:space:]]+Manage cluster-scoped BuildStrategies$'
  assert_line --regexp '^[[:space:]]+completion[[:space:]]+Generate the autocompletion script for the specified shell$'
  assert_line --regexp '^[[:space:]]+help[[:space:]]+Help about any command$'

}

@test "shp --help lists some Kubernetes flags" {
  run shp --help
  assert_success
  assert_line --regexp "--kubeconfig string[[:space:]]+Path to the kubeconfig file to use for CLI requests."
  assert_line --regexp "-n, --namespace string[[:space:]]+If present, the namespace scope for this CLI request"
  assert_line --regexp "--request-timeout string[[:space:]]+The length of time to wait before giving up on a single server request. Non-zero"
  refute_output --partial cache-dir
  refute_output --partial tls-server-name
}

@test "shp --help lists no logging flags" {
  run shp --help
  assert_success
  refute_output --partial log_dir
  refute_output --partial log_file
}

@test "shp [build/buildrun] create should not error when a name is specified" {
	# generate random names for our build and buildrun
  build_name=$(random_name)
  buildrun_name=$(random_name)

  # ensure that shp build create does not give an error when a build_name is specified
  run shp build create ${build_name} --source-git-url=url --output-image=image
  assert_success

   # ensure that shp buildrun create does not give an error when a buildrun_name is specified
  run shp buildrun create ${buildrun_name} --buildref-name=${build_name}
  assert_success
}

@test "shp -v=10 build list can log the kubernetes api communication" {
  # ensure that shp command doesn't log the api calls by default
  run shp build list
  assert_success
  refute_line --regexp "GET .*/apis/shipwright.io/v1beta1/namespaces/"
  refute_line --partial "Request Body"
  refute_line --partial "Response Body"

  # ensure that shp command supports -v loglevel flag.
  run shp -v=10 build list
  assert_success
  assert_line --regexp "GET .*/apis/shipwright.io/v1beta1/namespaces/"
  assert_line --partial "Request Body"
  assert_line --partial "Response Body"
}

@test "shp clusterbuildstrategy list shows sample strategy" {
  # pick the first sample ClusterBuildStrategy name
  sample=$(kubectl get clusterbuildstrategies -o jsonpath='{.items[0].metadata.name}')
  [ -n "$sample" ]

  run shp clusterbuildstrategy list
  assert_success
  assert_line --regexp "^${sample}[[:space:]]"
}

@test "shp buildstrategy delete removes a namespaced copy" {
  ns="${TEST_NAMESPACE:-default}"
  copy="e2e-bs-copy"

  # clone first cluster strategy into namespace as BuildStrategy
  kubectl get clusterbuildstrategies -o json \
    | jq --arg name "$copy" --arg ns "$ns" \
      '.items[0]
       | .kind="BuildStrategy"
       | .metadata.namespace=$ns
       | .metadata.name=$name
       | del(.metadata.uid,.metadata.resourceVersion,.metadata.creationTimestamp,.metadata.managedFields)' \
    | kubectl apply -f -

  run shp buildstrategy list -n "$ns"
  assert_success
  assert_line --regexp "^${copy}[[:space:]]"

  run shp buildstrategy delete "$copy" -n "$ns"
  assert_success

  run shp buildstrategy list -n "$ns"
  refute_line --regexp "^${copy}[[:space:]]"
}
