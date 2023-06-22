package arch

var (
	blockList = []Arch{
		// Default platforms to block as we don't support mobile/web
		{GOOS: "android"},
		{GOOS: "ios"},
		{GOOS: "js"},
		// 2023-06-21 block mips64 due to issues with symbol relocation in the compiler
		// https://github.com/peter-mount/go-build/issues/1 & https://github.com/peter-mount/piweather.center/issues/1
		// due to https://github.com/golang/go/issues/58240
		{GOOS: "openbsd", GOARCH: "mips64"},
	}
)

func equals(a, b string) bool {
	return a == b || a == ""
}

// IsBlocked returns true if Arch is in our blockList
func (a Arch) IsBlocked() bool {
	for _, blockEntry := range blockList {
		if blockEntry.GOOS == a.GOOS && equals(blockEntry.GOARCH, a.GOARCH) && equals(blockEntry.GOARM, a.GOARM) {
			return true
		}
	}
	return false
}
