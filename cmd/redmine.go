package cmd

import redmine "github.com/mattn/go-redmine"

type redmineder interface {
	Trackers() ([]redmine.IdName, error)
	IssuePriorities() ([]redmine.IssuePriority, error)
	CreateIssue(issue redmine.Issue) (*redmine.Issue, error)
}
