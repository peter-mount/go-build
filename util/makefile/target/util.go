package target

import "github.com/peter-mount/go-build/util/makefile"

type echo struct {
	label  string
	format string
	args   []any
}

func (m *echo) Build(b makefile.Builder) makefile.Builder {
	b.Echo(m.label, m.format, m.args...)
	return b
}

func (b *builder) Echo(label, format string, args ...any) Builder {
	t := &echo{
		label:  label,
		format: format,
		args:   args,
	}
	b.target.Add(t.Build)
	return b
}

type line struct {
	format string
	args   []any
}

func (m *line) Build(b makefile.Builder) makefile.Builder {
	b.Line(m.format, m.args...)
	return b
}

func (b *builder) Line(format string, args ...any) Builder {
	t := &line{
		format: format,
		args:   args,
	}
	b.target.Add(t.Build)
	return b
}

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
