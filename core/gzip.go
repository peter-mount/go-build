package core

import (
	"compress/gzip"
	"io"
	"os"
)

type GZip struct {
	Encoder *Encoder `kernel:"inject"`
	Gzip    *string  `kernel:"flag,gzip,gzip file to destination"`
	Gunzip  *string  `kernel:"flag,gunzip,gunzip file to destination"`
}

func (s *GZip) Start() error {
	if *s.Gunzip != "" {
		return s.gunzip()
	}
	return nil
}

func (s *GZip) gunzip() error {
	srcF, err := os.Open(*s.Gunzip)
	if err != nil {
		return err
	}
	defer srcF.Close()

	gr, err := gzip.NewReader(srcF)
	if err != nil {
		return err
	}
	defer gr.Close()

	destF, err := os.Create(*s.Encoder.Dest)
	if err != nil {
		return err
	}
	defer destF.Close()

	_, err = io.Copy(destF, gr)

	return err
}

func (s *GZip) gzip() error {
	srcF, err := os.Open(*s.Gzip)
	if err != nil {
		return err
	}
	defer srcF.Close()

	destF, err := os.Create(*s.Encoder.Dest)
	if err != nil {
		return err
	}
	defer destF.Close()

	gw := gzip.NewWriter(destF)
	if err != nil {
		return err
	}
	defer gw.Close()

	_, err = io.Copy(gw, srcF)

	return err
}
