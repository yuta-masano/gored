package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	redmine "github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// addCmd represents the add command.
var addCmd = &cobra.Command{
	Use:   "add project",
	Short: `add a new issue`,
	Long: `Create a new issue on Redmine using your clipboard text allowing you to edit
the issue subject and description via your editor.
After that, send the added issue page's title and URL into your clipboard.`,
	RunE: runAdd,
}

// Flags
var (
	tracker  string
	priority string
)

func init() {
	RootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&tracker, "tracker", "t", "", "choose your tracker")
	addCmd.Flags().StringVarP(&priority, "priority", "p", "", "choose your priority")
}

func censor(clipboardText string) string {
	// clipboard が 2 行以下 = クリップボードにパスワードが入っている可能性ありとみなして
	// clipboardText は使わない。
	if len(strings.Split(clipboardText, "\n")) <= 2 {
		return ""
	}
	return clipboardText
}

func validateContents(before, after []byte) ([]string, error) {
	// see: TestValidateContents(t *testing.T)
	if bytes.Equal(before, after) {
		return nil, errors.New("edit aborted")
	}
	if bytes.Equal(bytes.TrimRight(before, "\n"), bytes.TrimRight(after, "\n")) {
		return nil, errors.New("no changed")
	}
	// 空文字は不許可です。
	if string(after) == "" {
		return nil, errors.New("canceled")
	}
	// Split は分割対象がないとそのまま対象がスライスの 0 番目の要素になる。
	lines := strings.Split(string(after), "\n")
	if len(lines) == 0 { // 起こり得るのか？
		return nil, errors.New("canceled")
	}
	return lines, nil
}

func issueFromEditor(content string) (*redmine.Issue, error) {
	tmpFile, err := ioutil.TempFile("", ".gored.")
	if err != nil {
		return nil, err
	}
	defer func() {
		if e1 := tmpFile.Close(); e1 != nil { // tmpFile はたぶん Close してから Remove した方がいい気がする。
			panic(e1)
		}
		if e2 := os.Remove(tmpFile.Name()); e2 != nil {
			panic(e2)
		}
	}()

	// FIXME: 2016-12-16
	// Open, Close と別の関数でファイル操作している。
	// たぶん、あまりよくない。
	err = tmpWrite(tmpFile, content)
	if err != nil {
		return nil, err
	}

	before, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, err
	}
	editor := getEditor()
	if err = run(editor, tmpFile.Name()); err != nil {
		return nil, err
	}
	after, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, err
	}

	lines, err := validateContents(before, after)
	if err != nil {
		return nil, err
	}

	var issue redmine.Issue
	if len(lines) == 1 {
		issue.Subject = lines[0]
	} else {
		issue.Subject, issue.Description = lines[0], strings.Join(lines[1:], "\n")
	}
	return &issue, nil
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

func sendClipboard(addedIssue *redmine.Issue) error {
	buff := new(bytes.Buffer)
	fmt.Fprintf(buff, " [%s #%d: %s - %s - Redmine]\n %s/issues/%d\n",
		addedIssue.Tracker.Name, addedIssue.Id, addedIssue.Subject, addedIssue.Project.Name,
		cfg.Endpoint, addedIssue.Id)
	return clipboard.WriteAll(buff.String())
}

func createIssue(clipboardText string) error {
	var issue *redmine.Issue

	issue, err := issueFromEditor(censor(clipboardText))
	if err != nil {
		return err
	}

	cl := redmine.NewClient(cfg.Endpoint, cfg.Apikey)
	trackers, err := cl.Trackers()
	if err != nil {
		return err
	}
	priorities, err := cl.IssuePriorities()
	if err != nil {
		return err
	}

	issue.ProjectId, issue.TrackerId, issue.PriorityId =
		viper.GetInt("ProjectID"), retriveTracker(trackers).Id, retrievePriority(priorities).Id
	addedIssue, err := cl.CreateIssue(*issue)
	if err != nil {
		return err
	}

	return sendClipboard(addedIssue)
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("specify a project to add a new issue")
	}
	projectID, err := cfg.getProjetID(args[0])
	if err != nil {
		return err
	}
	viper.Set("ProjectID", projectID)

	clipboardText, err := clipboard.ReadAll()
	if err != nil {
		return err
	}
	if err := createIssue(clipboardText); err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}
