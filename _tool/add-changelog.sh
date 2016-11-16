#!/bin/bash

# Fail on unset variables, command errors and pipe fail.
set -o nounset -o errexit -o pipefail

# Prevent commands misbehaving due to locale differences.
export LC_ALL=C LANG=C

# trap for `mktemp`
trap 'rm -f /tmp/tmp.*."${0##*/}"'         0        # EXIT
trap 'rm -f /tmp/tmp.*."${0##*/}"; exit 1' 1 2 3 15 # HUP QUIT INT TERM

NEW_TAG="$1"
CURRENT_TAG="$(git describe --always --dirty)"
CURRENT_CHANGELOG="$(git show origin/master:CHANGELOG)"
COMMIT_LOGS=$(git log "${CURRENT_TAG%%-*}..." \
	--format='    * %s'                   \
	--grep='([a-z]\+ #[0-9]\+'            |
	sed -e 's/\(['^$'\x01''-'$'\x7e'']\) \(['^$'\x01''-'$'\x7e'']\)/\1\2/g')
	# 上の sed は、「全角 全角」となっている文字列から半角スペースを
	# 取り除いている。
	# 2 行以上のコミットログの件名を一行で表示すると、
	# 余計な半角スペースが含まれてしまうので、それを取り除くため。
	# 以下を使った割と強引な方法。
	# - bash の $'...' 表記を使って ASCII コード以外 = 半角文字以外を表現。
	# - bash の文字列結合は単に文字列を隣接させるだけでよい。

NEW_CHANGELOG="$(mktemp --tmpdir=/tmp --suffix=".${0##*/}")"
{
	echo "$NEW_TAG ($(date +'%F'))"
	echo "$COMMIT_LOGS"
	echo
	echo "$CURRENT_CHANGELOG"
} > "$NEW_CHANGELOG"

befor="$(md5sum "$NEW_CHANGELOG")"
vi "$NEW_CHANGELOG" < $(tty) > $(tty)
after="$(md5sum "$NEW_CHANGELOG")"
if [ "$befor" = "$after" ]; then
	echo 'CHANGELOG dit not modified' >&2
	exit 1
fi

cp --force "$NEW_CHANGELOG" CHANGELOG
git add CHANGELOG

CLOSE_ISSUES="$(echo $COMMIT_LOGS            |
	grep --only-matching -E '[a-z]+ #[0-9]+' |
	sed -e 's/[a-z]\+/close/')"

COMMIT_MESSAGE="$(mktemp --tmpdir=/tmp --suffix=".${0##*/}")"
{
	echo "Release $NEW_TAG"
	echo
	echo "$CLOSE_ISSUES"
} > "$COMMIT_MESSAGE"

vi "$COMMIT_MESSAGE" < $(tty) > $(tty)
git commit --file "$COMMIT_MESSAGE"
