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

@test "shp buildrun gather collects BuildRun diagnostics" {
  build_name=$(random_name)
  buildrun_name=$(random_name)
  output_image=$(get_output_image build-e2e)

  run shp build create ${build_name} \
    --source-git-url=https://github.com/shipwright-io/sample-go \
    --source-context-dir=source-build \
    --output-image=${output_image}
  assert_success

  run shp buildrun create ${buildrun_name} --buildref-name=${build_name}
  assert_success

  # verify that logs exist
  run shp buildrun logs --follow ${buildrun_name}
  assert_success

  # verify gathering of buildrun diagnostics
  run shp buildrun gather ${buildrun_name} --output "${BATS_TEST_TMPDIR}"
  assert_success
  assert_output --partial "BuildRun diagnostics written to"

  gather_dir="${BATS_TEST_TMPDIR}/buildrun-${buildrun_name}-gather"
  assert_file_exist "${gather_dir}/buildrun.yaml"
  
  # Depending on the BUILDRUN_EXECUTOR the buildrun maybe executed via a TaskRun or a PipelineRun.
  if [ -f "${gather_dir}/taskrun.yaml" ]; then
    assert_file_exist "${gather_dir}/taskrun.yaml"
    assert_file_exist "${gather_dir}/pod.yaml"
  elif [ -f "${gather_dir}/pipelinerun.yaml" ]; then
    assert_file_exist "${gather_dir}/pipelinerun.yaml"
    run test -d "${gather_dir}/taskruns"
    assert_success
    run test -d "${gather_dir}/pods"
    assert_succes
  else 
    fail "Expected either taskrun.yaml or pipelinerun.yaml to be generated in ${gather_dir}"
  fi
  
  # ensure log files exist in logs/ directory
  run find "${gather_dir}/logs" -type f -name "*.log"
  assert_success
  refute_output ""

  # shp buildrun gather --archive writes diagnostics as a .tar.gz archive
  run shp buildrun gather ${buildrun_name} --output "${BATS_TEST_TMPDIR}/archive" --archive
  assert_success
  assert_file_exist "${BATS_TEST_TMPDIR}/archive/buildrun-${buildrun_name}-gather.tar.gz"
  # ensure the dir is deleted
  [ ! -d "${BATS_TEST_TMPDIR}/archive/buildrun-${buildrun_name}-gather" ]
  assert_success
}
