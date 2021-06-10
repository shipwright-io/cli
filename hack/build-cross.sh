#!/bin/bash

set -e

GOFLAGS=$1
outputDir=$2

rm -rf "${outputDir}"
mkdir -p "${outputDir}"

linuxArches=( "amd64" "arm64" )
darwinArches=( "amd64" )
windowsArches=( "amd64" )

for arch in "${linuxArches[@]}"; do
    echo "Compiling shp for linux/${arch}"
    env GOOS=linux GOARCH="${arch}" go build ${GOFLAGS} -o "${outputDir}/shp-linux-${arch}" ./cmd/shp/...
done

for arch in "${darwinArches[@]}"; do
    echo "Compiling shp for darwin/${arch}"
    env GOOS=darwin GOARCH="${arch}" go build ${GOFLAGS} -o "${outputDir}/shp-darwin-${arch}" ./cmd/shp/...
done

for arch in "${windowsArches[@]}"; do
    echo "Compiling shp for windows/${arch}"
    env GOOS=windows GOARCH="${arch}" go build ${GOFLAGS} -o "${outputDir}/shp-windows-${arch}.exe" ./cmd/shp/...
done
