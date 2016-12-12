#!/bin/bash

# Fail on unset variables, command errors and pipe fail.
set -o nounset -o errexit -o pipefail

# Prevent commands misbehaving due to locale differences.
export LC_ALL=C LANG=C

ALL_OS="$1"
ALL_ARCH="$2"
LD_FLAGS="$3"
PKG_DEST_DIR="$4"
BINARY="$5"

cnt=0
for os in $ALL_OS; do
	if [ "_$os" = '_windows' ]; then
		app_name="${BINARY}.exe"
	else
		app_name="$BINARY"
	fi
	for arch in $ALL_ARCH; do
		echo "build $PKG_DEST_DIR/${os}_${arch}/$app_name"
		GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -a -tags netgo \
			-installsuffix netgo -ldflags "$LD_FLAGS"               \
			-o "$PKG_DEST_DIR/${os}_${arch}/$app_name"              &
		(( (cnt += 1) % 4 == 0 )) && wait
	done;
done
wait
