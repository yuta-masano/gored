#!/bin/bash

# Fail on unset variables, command errors and pipe fail.
set -o nounset -o errexit -o pipefail

# Prevent commands misbehaving due to locale differences.
export LC_ALL=C LANG=C

SOURCES=$(grep --binary-files=without-match --recursive --files-with-match    \
	--exclude-dir=vendor --extended-regexp 'Copyright...[0-9]{4} Yuta MASANO' | # Copyright © {{.year}} {{author}} が記載されているファイル名を取得。
	grep --invert-match 'LICENSE' || :)                                         # LICENSE ファイルは除く。

[ -z "$SOURCES" ] && exit

echo 'NG: the following sources still have a copyright sentence' >&2            # LICENSE ファイル以外の Copyright 文は不許可。
for source in $SOURCES; do
	echo "**** $source ****"
	head --lines 3 "$source"
	echo
done
exit 1
