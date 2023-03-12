package file

import (
	"encoding/json"
	"strings"
)

type File struct {
	Name         string
	Size         int64
	LastModified int64
}

func (f File) String() string {
	b, err := json.MarshalIndent(f, " ", "")
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(string(b), "\n", "")
}
