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

func (j Job) String() string {
	b, err := json.MarshalIndent(j, " ", "")
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(string(b), "\n", "")
}
