#!/bin/bash

# Fail on unset variables, command errors and pipe fail.
set -o nounset -o errexit -o pipefail

# Prevent commands misbehaving due to locale differences.
export LC_ALL=C LANG=C

ALL_OS_ARCH="$1"
PKG_DEST_DIR="$2"

cd "$PKG_DEST_DIR"
for os_arch in $ALL_OS_ARCH; do
	if $(echo "$os_arch" | grep --quiet 'linux'); then # is linux os?
		tar zcvf "../${os_arch}.tar.gz" "$os_arch"     #   -> tar.gz
	else                                               # is win, mac os?
		zip -r  "../${os_arch}.zip" "$os_arch"         #   -> zip
	fi
done
