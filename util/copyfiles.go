package util

import (
	"io"
	"os"
)

func CopyToWriter(path string, w io.Writer) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

func CopyFromReader(r io.Reader, dstName string, info os.FileInfo) error {
	err := copyFromReader(r, dstName, info)
	if err == nil {
		t := info.ModTime()
		// Note: Don't care if this fails, e.g. some file systems do not allow this
		_ = os.Chtimes(dstName, t, t)
	}
	return err
}

func copyFromReader(r io.Reader, dstName string, info os.FileInfo) error {
	var mode os.FileMode
	if info == nil {
		mode = 0644
	} else {
		mode = info.Mode()
	}
	dstFile, err := os.OpenFile(dstName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, r)
	return err
}

func copyFileByName(path string, dstName string, info os.FileInfo) error {
	dstFile, err := os.OpenFile(dstName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()
	return CopyToWriter(path, dstFile)
}

func CopyFile(path string, dstName string, info os.FileInfo) error {
	err := copyFileByName(path, dstName, info)
	if err == nil {
		t := info.ModTime()
		// Note: Don't care if this fails, e.g. some file systems do not allow this
		_ = os.Chtimes(dstName, t, t)
	}
	return err
}
