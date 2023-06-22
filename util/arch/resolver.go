package arch

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"sort"
)

func GetArches() ([]Arch, error) {
	var buf bytes.Buffer
	cmd := exec.Command("go", "tool", "dist", "list", "-json")
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var arches []Arch
	if err := json.Unmarshal(buf.Bytes(), &arches); err != nil {
		return nil, err
	}

	sort.SliceStable(arches, func(i, j int) bool {
		a, b := arches[i], arches[j]

		if a.GOOS != b.GOOS {
			return a.GOOS < b.GOOS
		}

		return a.GOARCH < b.GOARCH
	})

	// Filter out blocked platforms
	var a []Arch
	for _, e := range arches {
		if !e.IsBlocked() {
			if e.GOARCH == "arm" {
				// We support arm 6 & 7 for 32bits
				e.GOARM = "6"
				a = append(a, e)

				e.GOARM = "7"
				a = append(a, e)
			} else {
				a = append(a, e)
			}
		}
	}
	return a, nil
}
