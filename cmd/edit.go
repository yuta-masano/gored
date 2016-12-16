package cmd

import (
	"html/template"
	"os"
	"os/exec"
	"runtime"
)

const defaultContent string = `### Single Line Subject ###
### Start Description ###`

func getEditor() string {
	editor := cfg.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
		if editor == "" {
			if runtime.GOOS == windows {
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
	if runtime.GOOS == windows {
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

func tmpWrite(tmp *os.File, content string) error {
	if content == "" {
		_, err := tmp.Write([]byte(defaultContent))
		if err != nil {
			return err
		}
	} else {
		tpl := template.Must(template.New("").Parse(cfg.Template))
		err := tpl.Execute(tmp, struct{ Clipboard string }{Clipboard: content})
		if err != nil {
			return err
		}
	}

	return nil
}