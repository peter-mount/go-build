package target

import (
	"fmt"
	"github.com/peter-mount/go-build/util/makefile"
	"testing"
)

func Test_buildTool_Build(t *testing.T) {

	got := New().
		Target("builds/test").
		BuildTool("-test", "arg1", "arg2").
		Target("builds/dependency").
		MkDir("/tmp/some/path").
		BuildTool("-copy", "arg1", "arg2").
		Build(makefile.New()).
		Build()

	fmt.Println(got)
}
