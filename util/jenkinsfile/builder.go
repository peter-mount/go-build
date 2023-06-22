package jenkinsfile

import (
	"fmt"
	"sort"
	"strings"
)

type Builder interface {
	Begin(string, ...any) Builder
	Array() Builder
	Separator(string) Builder
	Line(string, ...any) Builder
	Property(string, any) Builder
	End() Builder
	Build() string

	Simple(string, ...string) Builder

	Node(string) Builder
	Stage(string) Builder
	Sh(string, ...any) Builder
	Parallel() Builder

	Sort() Builder
}

type builder struct {
	parent     *builder   // parent builder
	key        string     // Used for sorting stages
	parallel   bool       // Used for parallel blocks
	start      string     // Start of block, e.g. node("")
	separator  string     // Separator between lines & children, e.g. ","
	children   []*builder // Child builders
	terminator string     // Terminator e.g. "}" or "])"
}

func New() Builder {
	return &builder{}
}

func (b *builder) Array() Builder {
	return b.Separator(",")
}

func (b *builder) Separator(s string) Builder {
	b.separator = s
	return b
}

func (b *builder) Line(f string, a ...any) Builder {
	c := &builder{
		parent: b,
		start:  fmt.Sprintf(f, a...),
	}
	b.children = append(b.children, c)
	return b
}

func (b *builder) Property(k string, a any) Builder {
	return b.Line("%s: '%v'", k, a)
}

func (b *builder) Begin(f string, a ...any) Builder {
	return b.begin("", f, a...)
}

func (b *builder) begin(key, f string, a ...any) Builder {
	s := fmt.Sprintf(f, a...)

	c := &builder{parent: b, start: s, key: key}

	switch {
	case strings.HasSuffix(s, "(["):
		c.terminator = "])"
	case strings.HasSuffix(s, "("):
		c.terminator = ")"
	case strings.HasSuffix(s, "{"):
		c.terminator = "}"
	}
	b.children = append(b.children, c)
	return c
}

func (b *builder) End() Builder {
	if b.parent == nil {
		return b
	}
	return b.parent
}

func (b *builder) build(nest int, a []string) []string {
	return b.buildInner(nest, true, a)
}

func (b *builder) buildInner(nest int, startEnd bool, a []string) []string {
	n := nest - 1
	if n < 0 {
		n = 0
	}

	var a0 []string

	if b.parallel && len(b.children) == 1 {
		// If parallel & a single child then replace entirely with the child
		a0 = b.children[0].buildInner(n, false, a0)
	} else {

		if startEnd && b.start != "" {
			a0 = append(a0, strings.Repeat("  ", n)+b.start)
		}

		var a1 []string
		prefix := strings.Repeat("  ", nest)

		if b.parallel {
			for _, c := range b.children {
				if c.key != "" {
					var a2 []string
					a2 = append(a2, prefix+c.key+": {")
					a2 = c.buildInner(nest+1, false, a2)
					a2 = append(a2, prefix+"}")
					a1 = append(a1, strings.Join(a2, "\n"))
				}
			}

		} else {
			for _, c := range b.children {
				a1 = c.build(nest+1, a1)
			}
		}

		if b.separator != "" {
			l := len(a1) - 1
			for i, e := range a1 {
				if i < l && !strings.HasSuffix(e, b.separator) {
					a1[i] = e + b.separator
				}
			}
		}
		a0 = append(a0, a1...)

		if startEnd && b.terminator != "" {
			a0 = append(a0, strings.Repeat("  ", n)+b.terminator)
		}
	}

	return append(a, strings.Join(a0, "\n"))
}

func (b *builder) Build() string {
	return strings.Join(b.build(0, []string{}), "\n")
}

func (b *builder) Node(s string) Builder {
	return b.Begin("node(%q) {", s)
}

func (b *builder) Stage(s string) Builder {
	return b.begin(s, "stage(%q) {", s)
}

func (b *builder) Sh(f string, a ...any) Builder {
	return b.Line("sh '"+f+"'", a...)
}

func (b *builder) Parallel() Builder {
	c := &builder{
		parent:     b,
		parallel:   true,
		start:      "parallel(",
		separator:  ",",
		terminator: ")",
	}
	b.children = append(b.children, c)
	return c
}

func (b *builder) Sort() Builder {
	sort.SliceStable(b.children, func(i, j int) bool {
		c1 := b.children[i].key
		c2 := b.children[j].key
		return c1 == "" || c2 == "" || c1 < c2
	})
	return b
}

func (b *builder) Simple(p string, v ...string) Builder {
	c := &builder{
		parent: b,
		start:  p + "(" + strings.Join(v, ",") + ")",
	}
	b.children = append(b.children, c)
	return b
}
