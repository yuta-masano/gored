#!/bin/bash

# Fail on unset variables, command errors and pipe fail.
set -o nounset -o errexit -o pipefail

# Prevent commands misbehaving due to locale differences.
export LC_ALL=C LANG=C

NEW_TAG="$1"
TAG_LIST="$(git describe --always --dirty)"

echo "$TAG_LIST" | grep --quiet "$NEW_TAG" && :
if [ $? -eq 0 ]; then
	echo "$NEW_TAG already exists" >&2
	exit 1
fi

# CHANGELOG を上から一行ずつ読み込んでリリース向けバージョンに該当する
# 変更履歴だけを取り出す。
IS_TARGET_TAG=false
BUFF=""
while IFS= read line; do
	echo "$line" | grep --quiet "$NEW_TAG" && :
	if [ $? -eq 0 ]; then
		IS_TARGET_TAG=true
		BUFF+="$line\n"
		continue
	fi

	echo "$line" | grep --quiet -E "^[0-9]+\.[0-9]+\.[0-9]+" && :
	if [ $? -eq 0 ] && ($IS_TARGET_TAG); then
		IS_TARGET_TAG=false
		continue
	fi

	if ($IS_TARGET_TAG); then
		BUFF+="$line\n"
		continue
	fi
done < ./CHANGELOG

CHANGES="$(echo "$BUFF" | sed -e 's/\\n\\n//' -e 's/\\n/\n/g')"
git tag -a "$NEW_TAG" -m "$CHANGES"
git push --tags
