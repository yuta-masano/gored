package cmd

import (
	"os"
	"testing"
)

func TestGetEditor(t *testing.T) {
	expect := "vi"

	editor := getEditor()
	if editor != expect {
		t.Errorf("output shoud be %s, but %s", expect, editor)
		t.Fail()
	}
}

func TestRun(t *testing.T) {
	editor := "touch"
	file := "/tmp/gored.test"
	if err := run(editor, file); err != nil {
		t.Fail()
	}
	defer func() {
		err := os.Remove(file)
		if err != nil {
			panic(err)
		}
	}()
}
