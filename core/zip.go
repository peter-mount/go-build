package core

import (
	"archive/zip"
	"errors"
	"flag"
	"github.com/peter-mount/go-build/util"
	"github.com/peter-mount/go-kernel/v2/log"
	"github.com/peter-mount/go-kernel/v2/util/walk"
	"os"
	"path/filepath"
	"strings"
)

type Zip struct {
	Encoder *Encoder `kernel:"inject"`
	Zip     *bool    `kernel:"flag,zip,zip"`
}

func (s *Zip) Start() error {
	if *s.Zip {
		args := flag.Args()
		switch len(args) {
		case 2:
			return s.zip(args[0], args[1])

		default:
			return errors.New("-tar archive src")
		}
	}
	return nil
}

func (s *Zip) zip(archive, dir string) error {

	// Check dir exists. If it doesn't then do nothing.
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	util.Label("DIST ZIP", "%s %s", archive, dir)

	err = os.MkdirAll(filepath.Dir(archive), 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	packageName := getEnv("BUILD_PACKAGE_NAME")

	return walk.NewPathWalker().
		Then(func(path string, info os.FileInfo) (err error) {
			if info.IsDir() {
				return nil
			}

			name := strings.ReplaceAll(path, dir, packageName)
			if info.IsDir() {
				name = name + "/"
			}

			if log.IsVerbose() {
				log.Println(name)
			}

			w, err := zw.CreateHeader(&zip.FileHeader{
				Name:     name,
				Method:   zip.Deflate,
				Modified: info.ModTime(),
			})
			if err != nil {
				return err
			}

			if !info.IsDir() {
				err = util.CopyToWriter(path, w)
			}

			return err
		}).
		Walk(dir)
}
