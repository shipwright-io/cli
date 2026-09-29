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
}

# assert_shp_upload_output asserts common parts of the `shp build upload` subcommand.
function assert_shp_upload_output() {
	assert_output --partial 'Creating a BuildRun for'
	assert_output --partial 'created!'
	assert_output --partial 'to the Build POD'
}

# assert_shp_upload_follow_output inspects the output for the expected contents when using --follow
# flag, it should match parts of the Paketo (Buildpacks) output.
function assert_shp_upload_follow_output() {
	assert_output --partial '===> DETECTING'
	assert_output --partial '===> BUILDING'
	assert_output --partial '===> EXPORTING'
}

@test "shp build upload" {
	build_name=$(random_name)

	output_image="registry.registry.svc.cluster.local:32222/shipwright-io/build-e2e"
	source_url="https://github.com/shipwright-io/sample-go"
	repo_dir="${BATS_TEST_TMPDIR}/sample-go"

	# creating a new golang build with a modified context-dir, the context-dir will be subject to a
	# change directory (cd) during the build process, so if the directory is not uploaded properly
	# the actual build will fail
	run shp build create ${build_name} \
		--source-git-url="${source_url}" \
		--source-context-dir="source-build" \
		--output-image="${output_image}" \
		--output-insecure=true
	assert_success

	# cloning the same repository used for the build in the test temporary directory, this is the
	# path uploaded to the build pod
	run git clone "${source_url}" "${repo_dir}"
	assert_success

	#
	# Test Cases
	#

	run shp build upload ${build_name} "${repo_dir}"
	assert_success
	assert_shp_upload_output

	# uploading a dummy directory, on which the build won't be able to switch to the context-dir, so
	# we can simulate a error, after the data is streamed
	run shp build upload ${build_name} "${BATS_TEST_TMPDIR}"
	assert_failure

	run shp build upload --follow ${build_name} "${repo_dir}"
	assert_success
	assert_shp_upload_output
	assert_shp_upload_follow_output

	run shp build upload -F ${build_name} "${repo_dir}"
	assert_success
	assert_shp_upload_output
	assert_shp_upload_follow_output
}

@test "shp build upload with step resources" {
	build_name=$(random_name)

	output_image="registry.registry.svc.cluster.local:32222/shipwright-io/build-e2e"
	source_url="https://github.com/shipwright-io/sample-go"
	repo_dir="${BATS_TEST_TMPDIR}/sample-go"

	run shp build create ${build_name} \
		--source-git-url="${source_url}" \
		--source-context-dir="source-build" \
		--output-image="${output_image}" \
		--output-insecure=true
	assert_success

	run git clone "${source_url}" "${repo_dir}"
	assert_success

	#
	# Test Cases
	#

	# uploading with a per-step resource override, the created BuildRun must carry it on its spec
	run shp build upload ${build_name} "${repo_dir}" \
		--step-resources=build-and-push=limits.memory=1Gi \
		--step-resources=build-and-push=requests.cpu=250m
	assert_success
	assert_shp_upload_output

	# the created BuildRun spec must contain the requested step resources
	run kubectl get buildruns.shipwright.io \
		-o jsonpath='{.items[0].spec.stepResources[0].resources.limits.memory}'
	assert_success
	assert_output '1Gi'

	run kubectl get buildruns.shipwright.io \
		-o jsonpath='{.items[0].spec.stepResources[0].resources.requests.cpu}'
	assert_success
	assert_output '250m'

	# the build pod's step container must carry the requested memory limit
	run kubectl get pods -l "build.shipwright.io/name=${build_name}" \
		-o jsonpath='{.items[0].spec.containers[?(@.name=="step-build-and-push")].resources.limits.memory}'
	assert_success
	assert_output '1Gi'

	# an invalid step-resources value must be rejected before anything is created
	run shp build upload ${build_name} "${repo_dir}" \
		--step-resources=build-and-push=limits.memory=not-a-quantity
	assert_failure
}

@test "shp build upload into an existing BuildRun" {
	build_name=$(random_name)
	buildrun_name=$(random_name)

	output_image="registry.registry.svc.cluster.local:32222/shipwright-io/build-e2e"
	source_url="https://github.com/shipwright-io/sample-go"
	repo_dir="${BATS_TEST_TMPDIR}/sample-go"

	run shp build create ${build_name} \
		--source-git-url="${source_url}" \
		--source-context-dir="source-build" \
		--output-image="${output_image}" \
		--output-insecure=true
	assert_success

	run git clone "${source_url}" "${repo_dir}"
	assert_success

	# a BuildRun created out-of-band, carrying fields the upload command does not expose as flags
	# (here, a local source so the streaming waiter container is present)
	cat <<-EOF | kubectl create -f -
	apiVersion: shipwright.io/v1beta1
	kind: BuildRun
	metadata:
	  name: ${buildrun_name}
	spec:
	  build:
	    name: ${build_name}
	  source:
	    type: Local
	    local:
	      name: local-copy
	EOF

	#
	# Test Cases
	#

	# combining --buildrun-name with a creation-only flag must be rejected
	run shp build upload ${build_name} "${repo_dir}" \
		--buildrun-name="${buildrun_name}" \
		--sa-name=builder
	assert_failure
	assert_output --partial 'buildrun-name'

	# the pod of a BuildRun created earlier is usually already running when the upload starts;
	# kubectl wait fails at once when no pod matches yet, so wait for the pod to exist first
	for _ in $(seq 1 60); do
		[ -n "$(kubectl get pods -l "buildrun.shipwright.io/name=${buildrun_name}" -o name)" ] && break
		sleep 2
	done
	run kubectl wait pods -l "buildrun.shipwright.io/name=${buildrun_name}" \
		--for=jsonpath='{.status.phase}'=Running --timeout=180s
	assert_success

	# streaming into the pre-created BuildRun, following logs to completion; a global flag such as
	# --namespace is accepted alongside --buildrun-name
	ns="${TEST_NAMESPACE:-default}"
	run shp build upload --follow ${build_name} "${repo_dir}" \
		--namespace="${ns}" \
		--buildrun-name="${buildrun_name}"
	assert_success
	assert_output --partial 'Streaming into existing BuildRun'
	assert_output --partial 'to the Build POD'
	assert_shp_upload_follow_output

	# no extra BuildRun should have been created; only the one we made by hand
	run kubectl get buildruns.shipwright.io -o jsonpath='{.items[*].metadata.name}'
	assert_success
	assert_output "${buildrun_name}"
}

@test "shp build upload with bundle" {
	build_name=$(random_name)

	output_image="$(get_output_image build-e2e)"
	source_url="https://github.com/shipwright-io/sample-go"
	repo_dir="${BATS_TEST_TMPDIR}/sample-go"

	# Verify that invalid prune options are not accepted
	run shp build create ${build_name} \
		--source-oci-artifact-image="$(get_output_image source-bundle):latest" \
		--source-oci-artifact-prune=Magic
	assert_failure
	assert_output --partial 'invalid argument'

	# Create straightforward Dockerfile based build of the Go sample repository and with a
	# source bundle image specified, to make this build use bundle upload rather than the
	# local source copy approach.
	#
	# Note: This will only work if the registry used for the source bundle image is reachable
	# from the local shp client.
	#
	run shp build create ${build_name} \
		--source-oci-artifact-image="$(get_output_image source-bundle):latest" \
		--source-oci-artifact-prune=AfterPull \
		--source-context-dir="docker-build" \
		--dockerfile=Dockerfile \
		--strategy-name=kaniko \
		--output-image="${output_image}" \
		--output-insecure=true
	assert_success

	# Sample repository to be used for the test
	#
	run git clone "${source_url}" "${repo_dir}"
	assert_success

	#
	# Test Case
	#
	run shp build upload ${build_name} "${repo_dir}"
	assert_success
	assert_output --partial 'Creating a BuildRun for'
	assert_output --partial 'created!'
	assert_output --partial 'Uploading local source'
}
