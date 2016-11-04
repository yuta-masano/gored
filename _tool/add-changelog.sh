#!/bin/bash

# Fail on unset variables, command errors and pipe fail.
set -o nounset -o errexit -o pipefail

# Prevent commands misbehaving due to locale differences.
export LC_ALL=C LANG=C

NEW_TAG=$(git rev-parse --abbrev-ref HEAD |
	grep --only-matching -E '[0-9]+\.[0-9]+\.[0-9]+')
CURRENT_TAG="$(git describe --always --dirty | sed -e 's/-.*//')"
CURRENT_CHANGELOG="$(git show HEAD:./CHANGELOG)"
COMMIT_LOGS=$(git log "${CURRENT_TAG}..." --format='    * %s' --grep='([a-z]\+ #[0-9]\+')

echo "$NEW_TAG ($(date +'%F'))" >CHANGELOG
echo "$COMMIT_LOGS" >>CHANGELOG
echo >> CHANGELOG
echo "$CURRENT_CHANGELOG" >>CHANGELOG
