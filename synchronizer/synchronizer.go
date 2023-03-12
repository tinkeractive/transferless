package synchronizer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/tinkeractive/transferless/file"
	"github.com/tinkeractive/transferless/job"
	"github.com/tinkeractive/transferless/transfer"
	"github.com/rclone/rclone/cmd"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/filter"
	"github.com/rclone/rclone/fs/operations"
)

func Sync(transferObj transfer.Transfer) error {
	for _, target := range transferObj.Job.Targets {
		err := Copy(transferObj, target)
		if err != nil {
			return err
		}
	}
	if transferObj.Job.Source.Delete {
		err := Delete(transferObj)
		if err != nil {
			return err
		}
	}
	return nil
}

func Delete(transferObj transfer.Transfer) error {
	sourcePath := path.Clean(path.Join(transferObj.Job.Source.Root, transferObj.File.Name))
	log.Println("deleting source path:", sourcePath)
	sourceRemotePath := fmt.Sprintf("%s:%s", transferObj.Job.Source.Remote, sourcePath)
	fs, srcFileName := cmd.NewFsFile(sourceRemotePath)
	ctx, err := NewContext()
	if err != nil {
		return err
	}
	fileObj, err := fs.NewObject(ctx, srcFileName)
	if err != nil {
		return err
	}
	return operations.DeleteFile(ctx, fileObj)
}

func Copy(transferObj transfer.Transfer, target job.JobTarget) error {
	sourcePath := path.Clean(path.Join(transferObj.Job.Source.Root, transferObj.File.Name))
	log.Println("source path:", sourcePath)
	targetPath, err := GetTargetPath(transferObj.File, target)
	if err != nil {
		return err
	}
	log.Println("target path:", targetPath)
	ctx, err := NewContext()
	if err != nil {
		return err
	}
	fsrc, err := fs.NewFs(ctx, fmt.Sprintf("%s:%s", transferObj.Job.Source.Remote, path.Dir(sourcePath)))
	if err != nil {
		return err
	}
	fdst, err := fs.NewFs(ctx, fmt.Sprintf("%s:%s", target.Remote, path.Dir(targetPath)))
	if err != nil {
		return err
	}
	return operations.CopyFile(filter.SetUseFilter(ctx, false), fdst, fsrc, path.Base(targetPath), path.Base(sourcePath))
}

func GetSourcePath(transferObj transfer.Transfer) string {
	result := ""
	result = path.Clean(path.Join(transferObj.Job.Source.Root, transferObj.File.Name))
	return result
}

func GetTargetPath(source file.File, target job.JobTarget) (string, error) {
	result := ""
	dir := path.Dir(source.Name)
	ext := path.Ext(source.Name)
	name := strings.TrimSuffix(path.Base(source.Name), ext)
	ext = strings.TrimPrefix(ext, ".")
	datetime := time.Unix(int64(source.LastModified), 0)
	tmpl, err := template.New("FileName").Parse(target.Pattern)
	if err != nil {
		return result, err
	}
	type FilePatternArgs struct {
		Dir, Name, Extension, Date, Time string
	}
	filePatternArgs := FilePatternArgs{
		dir,
		name,
		ext,
		datetime.Format(target.DateFormat),
		datetime.Format(target.TimeFormat),
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, filePatternArgs)
	if err != nil {
		return result, err
	}
	targetName := string(buf.Bytes())
	result = path.Clean(path.Join(target.Root, targetName))
	return result, nil
}

func NewContext() (context.Context, error) {
	fi, err := filter.NewFilter(nil)
	if err != nil {
		return context.Background(), err
	}
	return filter.ReplaceConfig(context.Background(), fi), nil
}
