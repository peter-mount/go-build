package target

import "github.com/peter-mount/go-build/util/makefile"

type mkdir string

func (m *mkdir) Build(b makefile.Builder) makefile.Builder {
	b.Line("@mkdir -p %s", *m)
	return b
}

func (b *builder) MkDir(path string) Builder {
	t := mkdir(path)
	b.target.Add(t.Build)
	return b
}
