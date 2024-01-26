package arch

import (
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

var (
	blockList = &BlockList{
		Block: []BlockArch{
			// Default platforms to block as we don't support mobile/web
			{OS: "android"},
			{OS: "ios"},
			{OS: "js"},
			// 2023-06-21 block mips64 due to issues with symbol relocation in the compiler
			// https://github.com/peter-mount/go-build/issues/1 & https://github.com/peter-mount/piweather.center/issues/1
			// due to https://github.com/golang/go/issues/58240
			{OS: "openbsd", Arch: "mips64"},
		},
	}
)

type BlockList struct {
	// Replace the built-in list
	Replace bool `yaml:"replace"`
	// Architectures to block
	Block []BlockArch `yaml:"block"`
	// Tools to block
	Tools map[string][]BlockArch `yaml:"tools"`
}

func (a *BlockList) Merge(b *BlockList) *BlockList {
	if a == nil || b.Replace {
		return b
	}

	a.Block = merge(a.Block, b.Block)
	if len(b.Tools) > 0 {
		if len(a.Tools) == 0 {
			a.Tools = make(map[string][]BlockArch)
		}
		for k, e := range b.Tools {
			a.Tools[k] = merge(a.Tools[k], e)
		}
	}
	return a
}

func contains(a []BlockArch, b BlockArch) bool {
	for _, e := range a {
		if e.Equals(b) {
			return true
		}
	}
	return false
}

func merge(a, b []BlockArch) []BlockArch {
	for _, e := range b {
		if !contains(a, e) {
			a = append(a, e)
		}
	}
	return a
}

type BlockArch struct {
	OS   string `yaml:"os"`
	Arch string `yaml:"arch,omitempty"`
}

func (a BlockArch) Equals(b BlockArch) bool {
	return a.OS == b.OS && a.Arch == b.Arch
}

// IsBlocked returns true if Arch is in our blockList
func (a Arch) IsBlocked() bool {
	if blockList == nil {
		return false
	}
	return isArchBlocked(blockList.Block, a)
}

func (a Arch) IsToolBlocked(tool string) bool {
	if blockList == nil || blockList.Tools == nil {
		return false
	}

	toolArch, exists := blockList.Tools[strings.ToLower(tool)]
	if exists {
		return isArchBlocked(toolArch, a)
	}

	return false
}

func isArchBlocked(a []BlockArch, arch Arch) bool {
	for _, e := range a {
		if equals(e.OS, arch.GOOS) && equalsOptional(e.Arch, arch.GOARCH) {
			return true
		}
	}

	return false
}

func equals(a, b string) bool {
	return strings.ToLower(a) == strings.ToLower(b)
}

func equalsOptional(a, b string) bool {
	return a == "" || equals(a, b)
}

func LoadBlockList(n string) error {
	b, err := os.ReadFile(n)
	if err != nil {
		return err
	}

	bl := &BlockList{}
	err = yaml.Unmarshal(b, bl)
	if err != nil {
		return err
	}

	blockList = blockList.Merge(bl)

	return nil
}
