package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	redmine "github.com/yuta-masano/go-redmine"
)

var (
	version bool
	tracker string
	subject string

	description string
	priority    string

	projectID int

	// These values are embedded when building.
	buildVersion  string
	buildRevision string
	buildWith     string
)

var (
	trackerTable  = []string{"情報更新", "バグ", "機能", "サポート"}
	priorityTable = []string{"Low", "Normal", "High"}
)

type config struct {
	Endpoint string
	Apikey   string
	Editor   string
	Projects map[int]string
	Template string
}

var cfg config

var rootCmd = &cobra.Command{
	Use: "gored project_alias",
	Short: `gored creates a new issue on Redmine using your clipboard text,
sends the added issue page's title and URL into your clipboard.`,
	RunE: runGored,
}

const win = "windows"

func init() {
	rootCmd.Flags().BoolVarP(&version, "version", "v", false,
		"show program's version number and exit")
	rootCmd.Flags().StringVarP(&tracker, "tracker", "t", "バグ",
		fmt.Sprint("choose ", strings.Join(trackerTable, ", ")))
	rootCmd.Flags().StringVarP(&priority, "priority", "p", "Normal",
		fmt.Sprint("choose ", strings.Join(priorityTable, ", ")))
	rootCmd.Flags().BoolP("help", "h", false, "help for gored")
}

func runGored(cmd *cobra.Command, args []string) error {
	if version {
		fmt.Printf("version: %s\nrevision: %s\nwith: %s\n",
			buildVersion, buildRevision, buildWith)
		return nil
	}
	if len(args) < 1 {
		return errors.New("specify project_alias to add a new issue")
	}
	if !contain(trackerTable, tracker) {
		return fmt.Errorf("%s is invalid tracker", tracker)
	}
	if !contain(priorityTable, priority) {
		return fmt.Errorf("%s is invalid priority", priority)
	}

	if err := readConfig(); err != nil {
		return err
	}
	projectAlias := args[0]
	for k, v := range cfg.Projects {
		if v == projectAlias {
			projectID = k
		}
	}
	if projectID == 0 { // project_id は 1 から始まる（と思われる）。
		return fmt.Errorf("%s is invalid project_alias", args[0])
	}

	rand.Seed(time.Now().UnixNano())
	if err := createIssue(); err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func contain(haystack []string, needle string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}
	return false
}

func readConfig() error {
	var configDir string
	if runtime.GOOS == win {
		configDir = filepath.Join(os.Getenv("APPDATA"), "gored")
	} else {
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "gored")
	}
	if err := mkdir(configDir, 0700); err != nil {
		return err
	}
	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed in reading config file: %s", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed in setting config parameters: %s", err)
	}
	for _, param := range []string{"Endpoint", "Apikey"} {
		if !viper.IsSet(param) {
			return fmt.Errorf("failed in reading config parameter: %s must be specified", param)
		}
	}
	return nil
}

func mkdir(dir string, permission os.FileMode) error {
	finfo, err := os.Stat(dir)
	if err != nil {
		err = os.Mkdir(dir, permission)
		if err != nil {
			return err
		}
	} else if !finfo.IsDir() {
		return fmt.Errorf("%s mast be directory", dir)
	}
	return nil
}

func createIssue() error {
	var err error
	var issue *redmine.Issue

	clipboardText, err := clipboard.ReadAll()
	if err != nil {
		return err
	}
	// clipboard が 2 行以下 = クリップボードにパスワードが入っている可能性がありとみなして
	// clipboardText は使わない。
	if len(strings.Split(clipboardText, "\n")) <= 2 {
		issue, err = issueFromEditor("")
		if err != nil {
			return err
		}
	} else {
		issue, err = issueFromEditor(clipboardText)
		if err != nil {
			return err
		}
	}
	c := redmine.NewClient(cfg.Endpoint, cfg.Apikey)
	trackers, err := c.Trackers()
	if err != nil {
		return err
	}
	priorities, err := c.IssuePriorities()
	if err != nil {
		return err
	}

	issue.ProjectId = projectID
	issue.TrackerId = retriveTracker(trackers).Id
	issue.PriorityId = retrievePriority(priorities).Id
	addedIssue, err := c.CreateIssue(*issue)
	if err != nil {
		return err
	}
	err = sendClipboard(addedIssue)
	if err != nil {
		return err
	}
	return nil
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

func issueFromEditor(contents string) (*redmine.Issue, error) {
	file, err := ioutil.TempFile("", ".gored.")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = os.Remove(file.Name()); err != nil {
			panic(err)
		}
	}()

	tpl := template.Must(template.New("").Parse(cfg.Template))

	editor := getEditor()
	if contents == "" {
		contents = `### Single Line Subject ###
### Start Description ###
`
		_, err = file.Write([]byte(contents))
		if err != nil {
			return nil, err
		}
	} else {
		err := tpl.Execute(file, struct{ Clipboard string }{Clipboard: contents})
		if err != nil {
			return nil, err
		}

	}
	defer func() {
		if err = file.Close(); err != nil {
			panic(err)
		}
	}()
	if err = run(editor, file.Name()); err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(file.Name())
	if err != nil {
		return nil, err
	}

	text := string(b)
	if text == contents {
		return nil, errors.New("Canceled")
	}
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return nil, errors.New("Canceled")
	}

	if len(lines) == 1 {
		subject = lines[0]
	} else {
		subject, description = lines[0], strings.Join(lines[1:], "\n")
	}
	var issue redmine.Issue
	issue.Subject = subject
	issue.Description = description

	return &issue, nil
}

func getEditor() string {
	editor := cfg.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
		if editor == "" {
			if runtime.GOOS == win {
				editor = "notepad"
			} else {
				editor = "vi"
			}
		}
	}
	return editor
}

func run(editor, file string) error {
	cmd, err := exec.LookPath(editor)
	if err != nil {
		return err
	}
	// [linux - Trying to launch an external editor from within a Go program - Stack Overflow]
	// http://stackoverflow.com/questions/12088138/trying-to-launch-an-external-editor-from-within-a-go-program/12089980#12089980
	editorCmd := exec.Command(cmd, file)
	var stdin *os.File
	if runtime.GOOS == win {
		stdin, _ = os.Open("CONIN$")
	} else {
		stdin = os.Stdin
	}
	editorCmd.Stdin, editorCmd.Stdout, editorCmd.Stderr = stdin, os.Stdout, os.Stderr
	if err := editorCmd.Start(); err != nil {
		return err
	}
	if err := editorCmd.Wait(); err != nil {
		return err
	}
	return nil
}

func sendClipboard(addedIssue *redmine.Issue) error {
	buff := new(bytes.Buffer)
	fmt.Fprintf(buff, " [%s #%d: %s - %s - Redmine]\n %s/issues/%d\n",
		addedIssue.Tracker.Name, addedIssue.Id, addedIssue.Subject, addedIssue.Project.Name,
		cfg.Endpoint, addedIssue.Id)
	if err := clipboard.WriteAll(buff.String()); err != nil {
		return err
	}
	return nil
}
