package cmd

import (
	"reflect"
	"testing"

	"github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

func TestCensor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input  string
		expect string
	}{
		{
			input:  "foo",
			expect: "",
		},
		{
			input:  "foo\n",
			expect: "",
		},
		{
			input:  "foo\nbar",
			expect: "",
		},
		{
			input:  "foo\nbar\n",
			expect: "foo\nbar\n",
		},
	}

	for _, test := range testCases {
		output := censor(test.input)
		if output != test.expect {
			t.Fatalf("expect=%s, but got=%s", test.expect, output)
		}
	}
}

// go-redmine のモック。
type fakeRedmineCL struct {
	redmineder
}

func (f fakeRedmineCL) Trackers() ([]redmine.IdName, error) {
	return []redmine.IdName{{Id: 1, Name: "tracker1"}}, nil
}

func (f fakeRedmineCL) IssuePriorities() ([]redmine.IssuePriority, error) {
	return []redmine.IssuePriority{{Id: 11, Name: "priority0"}}, nil
}

func (f fakeRedmineCL) CreateIssue(issue redmine.Issue) (*redmine.Issue, error) {
	return &issue, nil
}

func TestCreateIssue(t *testing.T) {
	t.Parallel()

	tracker = "tracker1"
	priority = "priority0"
	viper.Set("ProjectID", 111)
	cl := new(fakeRedmineCL)

	testCases := []struct {
		input  []string
		expect *redmine.Issue
	}{
		{
			input:  []string{"foo"},
			expect: &redmine.Issue{Subject: "foo", ProjectId: 111, TrackerId: 1, PriorityId: 11},
		},
		{
			input:  []string{"foo", ""},
			expect: &redmine.Issue{Subject: "foo", ProjectId: 111, TrackerId: 1, PriorityId: 11},
		},
		{
			input:  []string{"foo", "bar", ""},
			expect: &redmine.Issue{Subject: "foo", Description: "bar\n", ProjectId: 111, TrackerId: 1, PriorityId: 11},
		},
	}

	for _, test := range testCases {
		issue, _ := createIssue(test.input, cl)
		if !reflect.DeepEqual(test.expect, issue) {
			t.Fatalf("expect=%v, but got=%v", test.expect, issue)
		}
	}
}
