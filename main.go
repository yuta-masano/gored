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
	redmine "github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version bool
	tracker string
	subject string

	description string
	priority    string

	projectID int

	// These values are embedded when building.
	buildVersion string
	buildDate    string
	buildWith    string
)

var cfg config

var (
	trackerTable  = []string{"情報更新", "バグ", "機能", "サポート"}
	priorityTable = []string{"Low", "Normal", "High"}
)

type config struct {
	Endpoint string
	Apikey   string
	Editor   string
	Projects map[int]string
}

type sendClipboardData struct {
	Issue *redmine.Issue
	Cfg   config
}

const sendClipboardText = ` [{{.Issue.Tracker.Name}} #{{.Issue.ID}} - {{.Issue.Project.Name}} - Redmine]
 {{.Cfg.Endpoint}}/issues/{{.Issue.ID}}`

var rootCmd = &cobra.Command{
	Use: "gored project_identifier",
	Short: `gored creates a new issue on Redmine using your clipboard text,
sends the added issue pages's title and URL into your clipboard.`,
	RunE: runGored,
}

var t = template.Must(template.New("").Parse(sendClipboardText))

const win = "windows"

func init() {
	rootCmd.Flags().BoolVarP(&version, "version", "v", false,
		"show program's version number and exit")
	rootCmd.Flags().StringVarP(&tracker, "tracker", "t", "バグ",
		fmt.Sprint("choose ", strings.Join(trackerTable, ", ")))
	rootCmd.Flags().StringVarP(&priority, "priority", "p", "Normal",
		fmt.Sprint("choose ", strings.Join(priorityTable, ", ")))
}

func runGored(cmd *cobra.Command, argv []string) error {
	if version {
		fmt.Printf("version: %s\nbuild at: %s\nwith: %s\n",
			buildVersion, buildDate, buildWith)
		return nil
	}
	if len(argv) < 1 {
		return errors.New("specify project_identifier to add a new issue\n")
	}
	if !contain(trackerTable, tracker) {
		return fmt.Errorf("%s is invalid tracker\n", tracker)
	}
	if !contain(priorityTable, priority) {
		return fmt.Errorf("%s is invalid priority\n", priority)
	}

	if err := readConfig(); err != nil {
		return err
	}
	projectIdentifier := argv[0]
	for k, v := range cfg.Projects {
		if v == projectIdentifier {
			projectID = k
		}
	}
	if projectID == 0 { // project_id は 1 から始まる（と思われる）。
		return fmt.Errorf("%s is invalid project_identifier\n", argv[0])
	}

	rand.Seed(time.Now().UnixNano())
	if err := createIssue(); err != nil {
		return fmt.Errorf("%s\n", err)
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
		return fmt.Errorf("failed in reading config file: %s\n", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed in setting config parameters: %s\n", err)
	}
	for _, param := range []string{"Endpoint", "Apikey"} {
		if !viper.IsSet(param) {
			return fmt.Errorf("failed in reading config parameter: %s must be specified\n", param)
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
		return fmt.Errorf("%s mast be directory\n", dir)
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
	issue.ProjectId = projectID
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

	editor := getEditor()
	if contents == "" {
		contents = `### Single Line Subject ###
### Start Description ###
`
	}
	_, err = file.Write([]byte(contents))
	if err != nil {
		return nil, err
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
	if err := t.Execute(buff, sendClipboardData{
		Issue: addedIssue,
		Cfg:   cfg,
	}); err != nil {
		return err
	}
	if err := clipboard.WriteAll(buff.String()); err != nil {
		return err
	}
	return nil
}
