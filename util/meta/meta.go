package meta

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Meta struct {
	CurrentDir    string // Current working directory
	ToolName      string // Path to tool name
	PackageName   string
	PackagePrefix string // package prefix from module line in go.mod
	Time          string // Time of build
	Uid           string // Userid or "N/A" if not available
	Version       string
}

func New() (*Meta, error) {
	m := &Meta{
		Time: time.Now().Format(time.RFC3339),
	}

	m.getUid()

	s, err := os.Getwd()
	if err == nil {
		m.CurrentDir = s
		m.ToolName = filepath.Join(m.CurrentDir, os.Args[0])

		err = m.getPrefix()
	}

	if err == nil {
		m.PackageName = filepath.Base(m.PackagePrefix)

		err = m.getVersion()
	}

	if err != nil {
		return nil, err
	}
	return m, nil
}

// extract module from go.mod
func (m *Meta) getPrefix() error {
	b, err := os.ReadFile("go.mod")
	if err != nil {
		return err
	}

	for _, s := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(s, "module") {
			a := strings.Split(s, " ")
			if len(a) == 2 {
				m.PackagePrefix = a[1]
				return nil
			}
		}
	}

	return errors.New("unable to read module from go.mod")
}

func runCmd(name string, args ...string) (string, error) {
	var buf bytes.Buffer

	cmd := exec.Command(name, args...)
	cmd.Stdout = &buf

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(buf.String(), "\n", ""), nil
}

// getUid returns the username/userid of the user running the command.
// Returns "N/A" if it cannot do this
func (m *Meta) getUid() {
	s, err := runCmd("id", "-u", "-n")
	if err == nil {
		m.Uid = s
	} else {
		m.Uid = "N/A"
	}
}

// getVersion returns the VERSION environment variable or the tag/commit from the git repository
func (m *Meta) getVersion() error {
	m.Version = os.Getenv("VERSION")

	if m.Version == "" {
		s, err := runCmd("git", "describe", "--tags", "--always", "--dirty", "--match=v*")
		if err != nil {
			return err
		}
		m.Version = strings.ReplaceAll(s, "-", ".")
	}

	return nil
}
