package job

import (
	"encoding/json"
	"strings"
)

type JobSource struct {
	Remote  string
	Root    string
	Pattern string
	Delete  bool `json:",omitempty"`
}

type JobTarget struct {
	Remote     string
	Root       string
	Pattern    string
	DateFormat string
	TimeFormat string
}

type Job struct {
	Name    string
	Source  JobSource
	Targets []JobTarget
}

// NOTE ini sections cannot have forward slash in the name
func (j *Job) Clean() error {
	var err error
	j.Source.Remote = strings.ReplaceAll(j.Source.Remote, "/", "")
	for i, target := range j.Targets {
		j.Targets[i].Remote = strings.ReplaceAll(target.Remote, "/", "")
	}
	return err
}

func (j Job) String() string {
	b, err := json.MarshalIndent(j, " ", "")
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(string(b), "\n", "")
}
