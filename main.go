package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	redmine "github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
)

var (
	version     bool
	projectID   string
	tracker     string
	subject     string
	description string
	priority    string

	// These values are embedded when building.
	buildVersion string
	buildDate    string
	buildWith    string
)

var (
	trackerTable  = []string{"情報更新", "バグ", "機能", "サポート"}
	priorityTable = []string{"Low", "Normal", "High"}
)

var rootCmd = &cobra.Command{
	Use: "gored project_id",
	Short: `gored adds a new issue using your clipboard text,
returns the added issue pages's title and URL.`,
	RunE: runGored,
}

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
		return fmt.Errorf("specify project_id to add a new issue\n")
	}
	if !contain(trackerTable, tracker) {
		return fmt.Errorf("%s is invalid tracker\n", tracker)
	}
	if !contain(priorityTable, priority) {
		return fmt.Errorf("%s is invalid priority\n", priority)
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

func createIssue() error {
	issue, err := issueFromEditor("")
	if err != nil {
		return err
	}
	spew.Dump(issue)
	// c := redmine.NewClient(conf.Endpoint, conf.Apikey)
	// issue.ProjectId = projectID
	// _, err = c.CreateIssue(*issue)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func issueFromEditor(contents string) (*redmine.Issue, error) {
	file, err := ioutil.TempFile("", ".gored.")
	if err != nil {
		return nil, err
	}
	defer os.Remove(file.Name())

	editor := getEditor()
	if contents == "" {
		contents = `### Single Line Subject  ###
### Start Description ###
`
	}
	file.Write([]byte(contents))
	defer file.Close()
	if err := run([]string{editor, file.Name()}); err != nil {
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
	var subject, description string
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
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vi"
		}
	}
	return editor
}

func run(args []string) error {
	cmd, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	var stdin *os.File
	if runtime.GOOS == "windows" {
		stdin, _ = os.Open("CONIN$")
	} else {
		stdin = os.Stdin
	}
	p, err := os.StartProcess(cmd, args, &os.ProcAttr{Files: []*os.File{
		stdin,
		os.Stdout,
		os.Stderr,
	}})
	if err != nil {
		return err
	}
	defer p.Release()
	w, err := p.Wait()
	if err != nil {
		return err
	}
	if !w.Exited() || !w.Success() {
		return errors.New("Failed to execute text editor")
	}
	return nil
}

func mkdir(dir string) error {
	finfo, err := os.Stat(dir)
	if err != nil { // err がある = no such file or directory のはず。
		os.Mkdir(dir, 0700)
	} else if !finfo.IsDir() {
		return fmt.Errorf("%s mast be directory", dir)
	}
	return nil
}
