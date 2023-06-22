package core

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"flag"
	"github.com/peter-mount/go-build/util"
	"github.com/peter-mount/go-kernel/v2/log"
	"github.com/peter-mount/go-kernel/v2/util/walk"
	"io"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

type Tar struct {
	Encoder *Encoder `kernel:"inject"`
	Tar     *bool    `kernel:"flag,tar,tar"`
}

func (s *Tar) Start() error {
	if *s.Tar {
		args := flag.Args()
		switch len(args) {
		case 2:
			return s.tar(args[0], args[1])

		default:
			return errors.New("-tar archive src")
		}
	}
	return nil
}

func (s *Tar) tar(archive, dir string) error {
	util.Label("DIST TAR", "%s %s", archive, dir)

	f, err := os.Create(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	packageName := getEnv("BUILD_PACKAGE_NAME")

	return walk.NewPathWalker().
		Then(func(path string, info os.FileInfo) (err error) {

			// get uid/gid, default to 0 if not supported
			var uid, gid int
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				uid = int(stat.Uid)
				gid = int(stat.Gid)
			}

			var userName, groupName string
			if user, err := user.LookupId(strconv.Itoa(uid)); err == nil && user != nil {
				userName = user.Name
			}

			if group, err := user.LookupGroupId(strconv.Itoa(gid)); err == nil && group != nil {
				groupName = group.Name
			}

			modTime := info.ModTime()
			name := strings.ReplaceAll(path, dir, packageName)

			if log.IsVerbose() {
				log.Println(name)
			}

			header := &tar.Header{
				Name:       name,
				Mode:       int64(info.Mode()),
				Uid:        uid,
				Gid:        gid,
				Uname:      userName,
				Gname:      groupName,
				ModTime:    modTime,
				AccessTime: modTime,
				ChangeTime: modTime,
			}

			if info.IsDir() {
				header.Typeflag = tar.TypeDir
				header.Name = header.Name + "/"
			} else {
				header.Typeflag = tar.TypeReg
				header.Size = info.Size()
			}

			err = tw.WriteHeader(header)
			if err != nil {
				return err
			}

			if !info.IsDir() {
				err = copyFile(path, tw)
			}

			return err
		}).
		Walk(dir)
}

func copyFile(path string, w io.Writer) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}
