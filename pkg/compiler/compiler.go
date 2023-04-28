package compiler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rclone/rclone/cmd"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/filter"
	"github.com/rclone/rclone/fs/operations"
	"github.com/tinkeractive/transferless/pkg/file"
	"github.com/tinkeractive/transferless/pkg/job"
)

func IsTransferCandidate(modTime, minModTime int64) bool {
	result := false
	if minModTime < modTime {
		result = true
	}
	return result
}

func PathMatchesRegex(objPath string, re *regexp.Regexp) bool {
	result := false
	if re.Match([]byte(objPath)) {
		result = true
	}
	return result
}

func IsLocked(remote, dataRoot, jobName string) (bool, error) {
	return GetMutexValue(remote, dataRoot, jobName)
}

func Lock(remote, dataRoot, jobName string) error {
	return PutMutex(remote, dataRoot, jobName, true)
}

func Unlock(remote, dataRoot, jobName string) error {
	return PutMutex(remote, dataRoot, jobName, false)
}

func GetMutexValue(remote, dataRoot, jobName string) (bool, error) {
	result := false
	ctx, err := NewContext()
	if err != nil {
		return result, err
	}
	var buf bytes.Buffer
	fsPath := fmt.Sprintf("%s:%s/mutex", remote, dataRoot)
	fsrc, err := fs.NewFs(ctx, fsPath)
	if err != nil {
		return result, err
	}
	fi := filter.GetConfig(ctx)
	err = fi.AddFile(jobName)
	if err != nil {
		return result, err
	}
	err = operations.Cat(ctx, fsrc, &buf, 0, -1)
	if err != nil {
		return result, err
	}
	val := buf.String()
	if val == "" {
		return result, nil
	}
	result, err = strconv.ParseBool(val)
	return result, err
}

func PutMutex(remote, dataRoot, jobName string, lock bool) error {
	ctx, err := NewContext()
	if err != nil {
		return err
	}
	reader := strings.NewReader(strconv.FormatBool(lock))
	readerCloser := io.NopCloser(reader)
	fsPath := fmt.Sprintf("%s:%s/mutex", remote, dataRoot)
	fdst, err := fs.NewFs(ctx, fsPath)
	if err != nil {
		return err
	}
	_, err = operations.Rcat(ctx, fdst, jobName, readerCloser, time.Now())
	return err
}

func GetLastModTime(remote, dataRoot, jobName string) (int64, error) {
	// TODO consider default value that will be used on first run (-1 vs current time)
	result := int64(-1)
	ctx, err := NewContext()
	if err != nil {
		return result, err
	}
	var buf bytes.Buffer
	fsPath := fmt.Sprintf("%s:%s/modtime/%s", remote, dataRoot, jobName)
	fsrc, fileName := cmd.NewFsFile(fsPath)
	fi := filter.GetConfig(ctx)
	err = fi.AddFile(fileName)
	if err != nil {
		return result, err
	}
	err = operations.Cat(ctx, fsrc, &buf, 0, -1)
	if err != nil {
		return result, err
	}
	val := buf.String()
	if val == "" {
		return result, nil
	}
	result, err = strconv.ParseInt(strings.TrimSpace(val), 10, 64)
	return result, err
}

func PutModTime(remote, dataRoot, jobName string, modTime int64) error {
	ctx, err := NewContext()
	if err != nil {
		return err
	}
	reader := strings.NewReader(strconv.FormatInt(modTime, 10))
	readerCloser := io.NopCloser(reader)
	fsPath := fmt.Sprintf("%s:%s/modtime", remote, dataRoot)
	fsrc, err := fs.NewFs(ctx, fsPath)
	if err != nil {
		return err
	}
	_, err = operations.Rcat(ctx, fsrc, jobName, readerCloser, time.Now())
	return err
}

func Compile(transferJob job.Job, lastModTime int64) ([]file.File, error) {
	transfers := []file.File{}
	ctx, err := NewContext()
	if err != nil {
		return transfers, err
	}
	cleanRoot := path.Clean(transferJob.Source.Root)
	fsPath := fmt.Sprintf("%s:%s", transferJob.Source.Remote, cleanRoot)
	fsrc, err := fs.NewFs(ctx, fsPath)
	if err != nil {
		return transfers, err
	}
	re, err := regexp.Compile(transferJob.Source.Pattern)
	if err != nil {
		return transfers, err
	}
	operations.ListFn(ctx, fsrc, Filter(lastModTime, re, &transfers))
	//	listJSONOpt := operations.ListJSONOpt{NoModTime: false}
	//	err = operations.ListJSON(ctx, fsrc, transferJob.Source.Remote, &listJSONOpt, NewFilter(lastModTime, re, &transfers))
	//	if err != nil {
	//		return transfers, err
	//	}
	return transfers, nil
}

func NewContext() (context.Context, error) {
	fi, err := filter.NewFilter(nil)
	if err != nil {
		return context.Background(), err
	}
	return filter.ReplaceConfig(context.Background(), fi), nil
}

func Filter(lastModTime int64, re *regexp.Regexp, transfers *[]file.File) func(fs.Object) {
	return func(obj fs.Object) {
		if PathMatchesRegex(obj.String(), re) {
			modTime := obj.ModTime(context.Background()).Unix()
			if IsTransferCandidate(modTime, lastModTime) {
				f := file.File{obj.String(), obj.Size(), modTime}
				*transfers = append(*transfers, f)
			}
		}
	}
}

// func NewFilter(lastModTime int64, re *regexp.Regexp, transfers *[]file.File) (func(*operations.ListJSONItem) error) {
// 	return func(obj *operations.ListJSONItem) error {
// 		log.Println(obj)
// 		return nil
// 	}
// }
