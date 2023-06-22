package build

import "github.com/peter-mount/go-kernel/v2"

func init() {
	// This will ensure that if Encoder is referenced then the default services are always
	// present to all projects using the build environment
	kernel.Register(
		&Build{},
		&Go{},
		&Tar{},
		&Zip{},
	)
}
