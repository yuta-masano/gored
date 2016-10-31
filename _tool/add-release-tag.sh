#!/bin/bash

# Fail on unset variables, command errors and pipe fail.
set -o nounset -o errexit -o pipefail

# Prevent commands misbehaving due to locale differences.
export LC_ALL=C LANG=C

TAG="$1"
TAG_LIST="$(git describe --always --dirty)"

echo "$TAG_LIST" | grep --quiet "$TAG" && :
if [ $? -eq 0 ]; then
	echo "$TAG already exists" >&2
	exit 1
fi

IS_TARGET_TAG=false
BUFF=""
while IFS= read line; do
	echo "$line" | grep --quiet "$TAG" && :
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
done <./CHANGELOG

MESSAGE="$(echo "$BUFF" | sed -e 's/\\n\\n//' -e 's/\\n/\n/g')"
echo "git -a \"$TAG\" -m \"$MESSAGE\""
echo "git push --tags"
