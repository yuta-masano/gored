package cmd

import (
	"strings"
	"testing"
)

func TestSensitiveCensor(t *testing.T) {
	inputs := []string{"foo", "foo\n", "foo\nbar"}
	expect := ""

	for _, v := range inputs {
		output := censor(v)
		if output != expect {
			t.Errorf("output shoud be %s, but %s", expect, output)
			t.Fail()
		}
	}
}

func TestSafetyCensor(t *testing.T) {
	input := "foo\nbar\n"
	expect := "foo\nbar\n"

	output := censor(input)
	if output != expect {
		t.Errorf("output shoud be %s, but %s", expect, output)
		t.Fail()
	}
}

func TestValidateContents(t *testing.T) {
	type testCase struct {
		before []string
		after  []string
		expect []string
	}

	errCheck := func(test testCase) {
		for i := range test.expect {
			_, err := validateContents([]byte(test.before[i]), []byte(test.after[i]))
			if err.Error() != test.expect[i] {
				t.Errorf("outuput shoud be %s, but %s", test.expect[i], err)
				t.Fail()
			}
		}
	}

	// `:q!` した場合。
	// before == after であれば、編集が abort されたことになる。
	var caseAbort = testCase{
		before: []string{"foo", "foo\n"},
		after:  []string{"foo", "foo\n"},
		expect: []string{"edit aborted", "edit aborted"},
	}
	errCheck(caseAbort)

	// 中身を編集せずに `:wq` した場合。
	// vim で末尾に改行が無い状態で保存すると、改行が自動で追記される。
	// lasteol 問題は vim の仕様らしい。
	// https://github.com/vim-jp/issues/issues/152
	var caseNoChanged = testCase{
		before: []string{"foo", "foo\n", "foo\n\n"},
		after:  []string{"foo\n", "foo\n", "foo\n\n"},
		expect: []string{"no changed", "edit aborted", "edit aborted"},
	}
	errCheck(caseNoChanged)

	// 中身が空の場合。
	var caseEmpty = testCase{
		before: []string{"foo"},
		after:  []string{""},
		expect: []string{"canceled"},
	}
	errCheck(caseEmpty)

	// 中身がちゃんと渡される場合。
	var caseOK = testCase{
		before: []string{"foo", "foo\n"},
		after:  []string{"abc", "abc\n"},
		expect: []string{"abc", "abc\n"},
	}
	for i := range caseOK.expect {
		lines, err := validateContents([]byte(caseOK.before[i]), []byte(caseOK.after[i]))
		if err != nil {
			t.Errorf("outuput shoud be nil, but %s", err)
			t.Fail()
		}
		if text := strings.Join(lines, "\n"); text != caseOK.expect[i] {
			t.Errorf("outuput shoud be %s, but %s", caseOK.expect[i], text)
			t.Fail()
		}
	}
}
