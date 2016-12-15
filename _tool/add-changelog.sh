#!/bin/bash

# Fail on unset variables, command errors and pipe fail.
set -o nounset -o errexit -o pipefail

# Prevent commands misbehaving due to locale differences.
export LC_ALL=C LANG=C

# trap for `mktemp`
trap 'rm -f /tmp/tmp.*."${0##*/}"'         0        # EXIT
trap 'rm -f /tmp/tmp.*."${0##*/}"; exit 1' 1 2 3 15 # HUP QUIT INT TERM

#---  新しい CHANGELOG を作成する  ---------------------------------------------
# 既存の CHANGELOG　の先頭に新しい変更履歴を挿入したいので、
#   1. 空のファイルに変更したい情報を記入。
#   2. 既存の変更履歴を 1. に追記。
#   3. 1. のファイルを CHANGELOG としてコピー。
# という手法を取っている。
NEW_TAG="$1"
CURRENT_TAG="$(git describe --always --dirty)"
CURRENT_CHANGELOG="$(git show origin/master:CHANGELOG)"
COMMIT_LOGS=$(git log "${CURRENT_TAG%%-*}..."                                 \
	--format='    * %s'                                                       \
	--grep='([a-z]\+ #[0-9]\+'                                               |\
	sed -e 's/\([^'$'\x01''-'$'\x7e'']\) \([^'$'\x01''-'$'\x7e'']\)/\1\2/g')
	# 上の sed は、「全角 全角」となっている文字列から半角スペースを
	# 取り除いている。
	# 2 行以上のコミットログの件名を一行で表示すると、
	# 余計な半角スペースが含まれてしまうので、それを取り除くため。
	# 以下を使った割と強引な方法。
	# - bash の $'...' 表記を使って ASCII コード以外 = 半角文字以外を表現。
	# - bash の文字列結合は単に文字列を隣接させるだけでよい。

FEATURE_LOGS="$(echo "$COMMIT_LOGS" | grep '(feature #' || :)"
BUG_LOGS="$(echo "$COMMIT_LOGS" | grep '(bug #' || :)"
ENHANCEMENT_LOGS="$(echo "$COMMIT_LOGS" | grep '(enhancement #' || :)"
MISC_LOGS="$(echo "$COMMIT_LOGS" | grep '(misc #' || :)"

NEW_CHANGELOG="$(mktemp --tmpdir=/tmp --suffix=".${0##*/}")"
{
	echo '# Delete this line to accept this draft.'
	echo "$NEW_TAG ($(date +'%F'))"
	if [ -n "$FEATURE_LOGS" ]; then
		echo '  Feature'
		echo "${FEATURE_LOGS//'(feature #'/'(#'}"
	fi
	if [ -n "$BUG_LOGS" ]; then
		echo '  Bug'
		echo "${BUG_LOGS//'(bug #'/'(#'}"
	fi
	if [ -n "$ENHANCEMENT_LOGS" ]; then
		echo '  Enhancement'
		echo "${ENHANCEMENT_LOGS//'(enhancement #'/'(#'}"
	fi
	if [ -n "$MISC_LOGS" ]; then
		echo '  Misc'
		echo "${MISC_LOGS//'(misc #'/'(#'}"
	fi
	echo
	echo "$CURRENT_CHANGELOG"
} > "$NEW_CHANGELOG"

#---  CHANGELOG が正しく編集されているかチェック  ------------------------------
befor="$(md5sum "$NEW_CHANGELOG")"
vi "$NEW_CHANGELOG" < $(tty) > $(tty)
after="$(md5sum "$NEW_CHANGELOG")"
if [ "$befor" = "$after" ]; then
	echo 'CHANGELOG dit not modified' >&2
	exit 1
fi
grep --quiet '# Delete this line' "$NEW_CHANGELOG" && :
if [ $? -eq 0 ]; then
	echo '1 st line must be deleted' >&2
	exit 1
fi

#---  CHANGELOG を適用してコミット  --------------------------------------------
cp --force "$NEW_CHANGELOG" CHANGELOG
git add CHANGELOG

CLOSE_ISSUES="$(echo $COMMIT_LOGS            |\
	grep --only-matching -E '[a-z]+ #[0-9]+' |\
	sed -e 's/[a-z]\+/close/'                |\
	uniq)"

COMMIT_MESSAGE="$(mktemp --tmpdir=/tmp --suffix=".${0##*/}")"
{
	echo "Release $NEW_TAG"
	echo
	echo "$CLOSE_ISSUES"
} > "$COMMIT_MESSAGE"

vi "$COMMIT_MESSAGE" < $(tty) > $(tty)
git commit --file "$COMMIT_MESSAGE"
