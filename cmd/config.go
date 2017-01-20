package cmd

import "fmt"

type config struct {
	Endpoint   string
	Apikey     string
	Editor     string
	ProjectID  int
	Projects   map[int]string
	Trackers   []string
	Priorities []string
	Template   string
}

func (c *config) getProjetID(name string) (int, error) {
	for k, v := range c.Projects {
		if v == name {
			return k, nil
		}
	}
	// project_id は 1 から始まる（と思われる）。
	return 0, fmt.Errorf("%s is invalid project", name)
}

var cfg config

var cfgFilePath string
