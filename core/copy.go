package core

import (
	"github.com/peter-mount/go-kernel/v2/util/walk"
	"os"
	"path/filepath"
	"strings"
)

type Copy struct {
	Encoder  *Encoder `kernel:"inject"`
	CopyDir  *string  `kernel:"flag,copydir,copy directory to dest"`
	CopyFile *string  `kernel:"flag,copyfile,copy file to dest"`
}

func (s *Copy) Start() error {
	if *s.CopyFile != "" {
		return s.copyFile()
	}

	if *s.CopyDir != "" {
		return walk.NewPathWalker().
			Then(s.copyDir).
			Walk(*s.CopyDir)
	}

	return nil
}

func (s *Copy) copyFile() error {
	info, err := os.Stat(*s.CopyFile)
	if err != nil {
		return err
	}

	return copyFileWithTimes(*s.CopyFile, *s.Encoder.Dest, info)
}

func (s *Copy) copyDir(path string, info os.FileInfo) error {
	// Ignore the source base directory
	if path == *s.CopyDir {
		return nil
	}

	// dest is the source minus the source base directory name
	dstName := filepath.Join(*s.Encoder.Dest, strings.TrimPrefix(path, *s.CopyDir+"/"))

	if info.IsDir() {
		return os.MkdirAll(dstName, info.Mode())
	}

	return copyFileWithTimes(path, dstName, info)
}

func copyFileByName(path string, dstName string, info os.FileInfo) error {
	dstFile, err := os.OpenFile(dstName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()
	return copyFile(path, dstFile)
}

func copyFileWithTimes(path string, dstName string, info os.FileInfo) error {
	err := copyFileByName(path, dstName, info)
	if err == nil {
		t := info.ModTime()
		// Note: Don't care if this fails, e.g. some file systems do not allow this
		_ = os.Chtimes(dstName, t, t)
	}
	return err
}
