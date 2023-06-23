package core

import (
	"github.com/peter-mount/go-kernel/v2/util/walk"
	"os"
	"path/filepath"
	"strings"
)

type Copy struct {
	Encoder *Encoder `kernel:"inject"`
	Source  *string  `kernel:"flag,copy,copy file/directory to dest"`
}

func (s *Copy) Start() error {
	if *s.Source != "" {
		return walk.NewPathWalker().
			Then(s.copy).
			Walk(*s.Source)
	}
	return nil
}

func (s *Copy) copy(path string, info os.FileInfo) error {
	// Ignore the source base directory
	if path == *s.Source {
		return nil
	}

	// dest is the source minus the source base directory name
	dstName := filepath.Join(*s.Encoder.Dest, strings.TrimPrefix(path, *s.Source+"/"))

	if info.IsDir() {
		return os.MkdirAll(dstName, info.Mode())
	}

	dstFile, err := os.OpenFile(dstName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	return copyFile(path, dstFile)
}
