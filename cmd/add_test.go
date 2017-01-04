package cmd

import "testing"

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
