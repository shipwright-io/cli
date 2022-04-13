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
		--source-url="${source_url}" \
		--source-context-dir="source-build" \
		--output-image="${output_image}"
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

@test "shp build upload with bundle" {
	build_name=$(random_name)

	output_image="$(get_output_image build-e2e)"
	source_url="https://github.com/shipwright-io/sample-go"
	repo_dir="${BATS_TEST_TMPDIR}/sample-go"

	# Verify that invalid prune options are not accepted
	run shp build create ${build_name} \
		--source-bundle-image="$(get_output_image source-bundle):latest" \
		--source-bundle-prune=Magic
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
		--source-bundle-image="$(get_output_image source-bundle):latest" \
		--source-bundle-prune=AfterPull \
		--source-context-dir="docker-build" \
		--dockerfile=Dockerfile \
		--strategy-name=kaniko \
		--output-image="${output_image}"
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
