#!/usr/bin/env bash
#
# Verify that generated Markdown documentation fils are out of sync.
#

set -eu -o pipefail

if ( git diff --name-only |grep -q '^docs\/' ) ; then
	echo "[ERROR] Markdown documentation is out of sync, run 'make generate-docs' and commit the changes!"
	exit 1
fi
