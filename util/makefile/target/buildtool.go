package target

import (
	"github.com/peter-mount/go-build/util/makefile"
	"strings"
)

type buildTool struct {
	parent *Target
	flag   string   // Flag to use if invoking the build tool
	args   []string // Arguments when BuildFlag is in use
}

func (b *builder) BuildTool(flag string, args ...string) Builder {
	t := buildTool{
		parent: b.target,
		flag:   flag,
		args:   args,
	}
	b.target.Add(t.Build)
	return b
}

func (bt *buildTool) Build(b makefile.Builder) makefile.Builder {
	b.Echo("BUILD", "%s %s", bt.flag, bt.parent.Target()).
		Line("@$(BUILD) %s %s",
			bt.flag,
			strings.Join(bt.args, " "))
	return b
}
