package transfer

import (
	"encoding/json"
	"strings"

	"github.com/tinkeractive/transferless/file"
	"github.com/tinkeractive/transferless/job"
)

type Transfer struct {
	File file.File
	Job  job.Job
}

func (t Transfer) String() string {
	b, err := json.MarshalIndent(t, " ", "")
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(string(b), "\n", "")
}
