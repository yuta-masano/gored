package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yuta-masano/go-tempedit"
)

const defaultContent = `### on line subject
### description
`

// addCmd represents the add command.
var addCmd = &cobra.Command{
	Use:   "add project",
	Short: `add a new issue`,
	Long: `Create a new issue on Redmine using your clipboard text allowing you to edit
the issue subject and description via your editor.
And then, send the title and URL of added issue page into your clipboard.`,
	RunE: runAdd,
}

// Flags.
var (
	tracker  string
	priority string
)

func init() {
	RootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&tracker, "tracker", "t", "", "choose your tracker (defalut a first element of Trackers list in config file)")
	addCmd.Flags().StringVarP(&priority, "priority", "p", "", "choose your priority (defalut a first element of Priorities list in config file)")
}

func censor(clipboardText string) string {
	// clipboardText が 2 行以下 = クリップボードにパスワードが入っている可能性あり
	// とみなして clipboardText は使わない。
	if len(strings.Split(clipboardText, "\n")) <= 2 {
		return ""
	}
	return clipboardText
}

func contentFromEditor(content string) ([]string, error) {
	edit := tempedit.New(cfg.Editor)
	if err := edit.MakeTemp("", ".gored."); err != nil {
		return nil, err
	}
	defer edit.CleanTempFile()

	if content == "" {
		if err := edit.Write(defaultContent); err != nil {
			return nil, err
		}
	} else {
		err := edit.WriteTemplate(
			cfg.Template,
			struct{ Clipboard string }{Clipboard: content},
		)
		if err != nil {
			return nil, err
		}

	}
	if err := edit.Run(); err != nil {
		return nil, err
	}
	if changed, err := edit.FileChanged(); !changed {
		return nil, err
	}
	return edit.Line(), nil
}

func retriveTracker(trackers []redmine.IdName) *redmine.IdName {
	for _, t := range trackers {
		if t.Name == tracker {
			return &redmine.IdName{Id: t.Id, Name: t.Name}
		}
	}
	return new(redmine.IdName)
}

func retrievePriority(priorities []redmine.IssuePriority) *redmine.IdName {
	for _, p := range priorities {
		if p.Name == priority {
			return &redmine.IdName{Id: p.Id, Name: p.Name}
		}
	}
	return new(redmine.IdName)
}

func createIssue(issueTextLine []string, cl redmineder) (*redmine.Issue, error) {
	var issue redmine.Issue

	if len(issueTextLine) == 1 {
		issue.Subject = issueTextLine[0]
	} else {
		issue.Subject, issue.Description = issueTextLine[0], strings.Join(issueTextLine[1:], "\n")
	}
	trackers, err := cl.Trackers()
	if err != nil {
		return nil, err
	}
	priorities, err := cl.IssuePriorities()
	if err != nil {
		return nil, err
	}

	issue.ProjectId, issue.TrackerId, issue.PriorityId =
		viper.GetInt("ProjectID"), retriveTracker(trackers).Id, retrievePriority(priorities).Id
	addedIssue, err := cl.CreateIssue(issue)
	if err != nil {
		return nil, err
	}
	return addedIssue, nil
}

func sendClipboard(addedIssue *redmine.Issue) error {
	buff := new(bytes.Buffer)
	fmt.Fprintf(buff, "[%s #%d: %s - %s - Redmine]\n%s/issues/%d\n",
		addedIssue.Tracker.Name, addedIssue.Id, addedIssue.Subject, addedIssue.Project.Name,
		cfg.Endpoint, addedIssue.Id)
	return clipboard.WriteAll(buff.String())
}

func addIssue(clipboardText string) error {
	safeText := censor(clipboardText)
	issueTextLine, err := contentFromEditor(safeText)
	if err != nil {
		return err
	}
	cl := redmine.NewClient(cfg.Endpoint, cfg.Apikey)
	addedIssue, err := createIssue(issueTextLine, cl)
	if err != nil {
		return err
	}
	return sendClipboard(addedIssue)
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("too few argements: specify a project to add a new issue")
	}
	var err error
	prjID, err := cfg.getProjetID(args[0])
	viper.Set("ProjectID", prjID)
	if err != nil {
		return err
	}
	clipboardText, err := clipboard.ReadAll()
	if err != nil {
		return err
	}
	if err := addIssue(clipboardText); err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}
