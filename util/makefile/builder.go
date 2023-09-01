package makefile

import (
	"fmt"
	"strings"
)

type Builder interface {
	Blank() Builder
	Block() Builder
	Comment(string, ...any) Builder

	Command(string, ...any) Builder
	Include(string, ...any) Builder
	SetVar(string, string, ...any) Builder

	Line(string, ...any) Builder
	Rule(string, ...string) Builder
	Phony(...string) Builder

	Echo(string, string, ...any) Builder
	Mkdir(...string) Builder
	RM(...string) Builder

	End() Builder
	Build() string

	AddDependency(...string) Builder
	Add(Handler) Builder

	IsBlank() bool
	IsBlock() bool
	IsComment() bool
	IsCommand() bool
	IsLine() bool
	IsRule() bool
	IsEmptyRule() bool
	NumTargets() int
}

// Handler is a function that can add to a Builder
type Handler func(Builder) Builder

// Of creates a Handler
func Of(handlers ...Handler) Handler {
	var r Handler
	for _, h := range handlers {
		r = r.Then(h)
	}
	return r
}

// Then creates a handler that weill invoke this one then a second
// with the same builder. The builder is returned unchanged.
func (a Handler) Then(b Handler) Handler {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return func(builder Builder) Builder {
		_ = a(builder)
		_ = b(builder)
		return builder
	}
}

// Chain a handler so that this Handler is invoked, and then it's returned builder is passed to b
func (a Handler) Chain(b Handler) Handler {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return func(builder Builder) Builder {
		return b(a(builder))
	}
}

// Do invokes this Handler. If this Handler is nil then the builder is returned as-is
func (a Handler) Do(builder Builder) Builder {
	if a != nil {
		return a(builder)
	}
	return builder
}

type builder struct {
	parent   *builder   // parent builder
	block    bool       // true for a block
	blank    bool       // true for a blank line
	key      string     // Rule
	line     string     // line in rule
	comment  string     // comment
	command  string     // command, e.g. "include xxx.inc"
	children []*builder // Child builders
}

func New() Builder {
	return &builder{}
}

func (b *builder) Add(h Handler) Builder {
	if h != nil {
		_ = h(b)
	}
	return b
}

func (b *builder) Blank() Builder {
	c := &builder{
		parent: b,
		blank:  true,
	}
	b.children = append(b.children, c)
	return b
}

func (b *builder) IsBlank() bool {
	return b.blank
}

func (b *builder) Block() Builder {
	c := &builder{
		parent: b,
		block:  true,
	}
	b.children = append(b.children, c)
	return c
}

func (b *builder) IsBlock() bool {
	return b.block
}

func (b *builder) End() Builder {
	if b.parent == nil {
		return b
	}
	return b.parent
}

func (b *builder) Comment(f string, a ...any) Builder {
	c := &builder{
		parent:  b,
		comment: fmt.Sprintf(f, a...),
	}
	b.children = append(b.children, c)
	return b
}

func (b *builder) IsComment() bool {
	return b.comment != ""
}

func (b *builder) Command(f string, a ...any) Builder {
	c := &builder{
		parent:  b,
		command: fmt.Sprintf(f, a...),
	}
	b.children = append(b.children, c)
	return b
}

func (b *builder) IsCommand() bool {
	return b.command != ""
}

func (b *builder) Include(f string, a ...any) Builder {
	return b.Command("include "+f, a...)
}

func (b *builder) SetVar(name, f string, a ...any) Builder {
	return b.Command(name+" = "+f, a...)
}

func (b *builder) Phony(dependencies ...string) Builder {
	return b.Rule(".PHONY", dependencies...)
}

func (b *builder) Line(f string, a ...any) Builder {
	c := &builder{parent: b, line: fmt.Sprintf(f, a...)}
	b.children = append(b.children, c)
	return b
}

func (b *builder) Echo(n string, f string, a ...any) Builder {
	return b.Line(fmt.Sprintf(`@echo "%-8s %s";\`, n, fmt.Sprintf(f, a...)))
}

func (b *builder) Mkdir(dirs ...string) Builder {
	return b.Line("@mkdir -p %s", strings.Join(dirs, " "))
}

func (b *builder) RM(dirs ...string) Builder {
	return b.Echo("RM", strings.Join(dirs, " ")).
		Line("rm -rf %s", strings.Join(dirs, " "))
}

func (b *builder) IsLine() bool {
	return b.line != "" && !b.IsRule()
}

func (b *builder) Rule(rule string, dependencies ...string) Builder {
	// If we are adding a rule inside a rule, then add it as a dependency
	switch {
	// adding a rule to a rule, then add it as a dependency
	case b.IsRule():
		b.AddDependency(rule)

	// adding a rule to a block, then if the block is in a rule
	// add it as a dependency to the blocks parent rule
	case b.IsBlock() && b.parent != nil && b.parent.IsRule():
		b.parent.AddDependency(rule)
	}

	c := &builder{
		parent: b,
		key:    rule,
	}
	c.AddDependency(dependencies...)
	b.children = append(b.children, c)
	return c
}

func (b *builder) IsRule() bool {
	return b.key != ""
}

func (b *builder) IsEmptyRule() bool {
	return b.IsRule() && b.line == ""
}

func (b *builder) NumTargets() int {
	if b.IsRule() {
		return len(strings.Split(b.line, " "))
	}
	return 0
}

func (b *builder) AddDependency(dependencies ...string) Builder {
	if !b.IsRule() {
		panic("not a rule")
	}

	// Split existing dependencies
	deps := strings.Split(b.line, " ")

	m := map[string]bool{}
	for _, dep := range deps {
		if !strings.HasPrefix(dep, ".") {
			m[dep] = true
		}
	}

	// Add new ones only if they are not already present
	for _, dep := range dependencies {
		if !strings.HasPrefix(dep, ".") {
			if _, ok := m[dep]; !ok {
				deps = append(deps, dep)
			}
		}
	}

	b.line = strings.TrimSpace(strings.Join(deps, " "))
	return b
}

func (b *builder) Build() string {
	return strings.Join(b.build([]string{}), "\n")
}

func (b *builder) build(a []string) []string {
	switch {
	case b.IsBlank():
		a = append(a, "")

	case b.IsComment():
		a = append(a, "# "+b.comment)

	case b.IsCommand():
		a = append(a, b.command)

	case b.IsRule():
		a = append(a, "", b.key+": "+b.line)

	case b.IsLine():
		a = append(a, "\t"+b.line)
	}

	for _, c := range b.children {
		a = c.build(a)
	}

	return a
}
