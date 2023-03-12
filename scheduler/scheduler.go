package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/tinkeractive/transferless/job"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/filter"
	"github.com/rclone/rclone/fs/operations"
)

func GetJobs(remote, sourcePath string) ([]job.Job, error) {
	var jobs []job.Job
	ctx, err := NewContext()
	if err != nil {
		return jobs, err
	}
	fsPath := fmt.Sprintf("%s:%s", remote, filepath.Dir(sourcePath))
	fsrc, err := fs.NewFs(ctx, fsPath)
	if err != nil {
		return jobs, err
	}
	fi := filter.GetConfig(ctx)
	err = fi.AddFile(filepath.Base(sourcePath))
	if err != nil {
		return jobs, err
	}
	var buf bytes.Buffer
	err = operations.Cat(ctx, fsrc, &buf, 0, -1)
	if err != nil {
		return jobs, err
	}
	err = json.Unmarshal(buf.Bytes(), &jobs)
	return jobs, err
}

func NewContext() (context.Context, error) {
	fi, err := filter.NewFilter(nil)
	if err != nil {
		return context.Background(), err
	}
	return filter.ReplaceConfig(context.Background(), fi), nil
}
