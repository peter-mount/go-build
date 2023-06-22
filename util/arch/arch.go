package arch

import (
	"path/filepath"
	"strings"
)

// Arch output from go tool dist list
type Arch struct {
	GOOS         string `json:"GOOS"`
	GOARCH       string `json:"GOARCH"`
	GgoSupported bool   `json:"GgoSupported"`
	FirstClass   bool   `json:"FirstClass"`
	GOARM        string `json:"-"`
}

func (a Arch) IsMobile() bool {
	return a.GOOS == "android" || a.GOOS == "ios" || a.GOOS == "js"
}

func (a Arch) IsWindows() bool {
	return a.GOOS == "windows"
}

func (a Arch) Platform() string {
	return strings.Join([]string{a.GOOS, a.GOARCH, a.GOARM}, ":")
}

func (a Arch) Arch() string {
	return a.GOARCH + a.GOARM
}

func (a Arch) Target() string {
	return a.GOOS + "_" + a.Arch()
}

func (a Arch) BaseDir(builds string) string {
	return filepath.Join(builds, a.GOOS, a.Arch())
}

func (a Arch) Tool(builds, tool string) string {
	if a.GOOS == "windows" {
		tool = tool + ".exe"
	}
	return filepath.Join(a.BaseDir(builds), "bin", tool)
}
