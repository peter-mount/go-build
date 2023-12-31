package target

import "github.com/peter-mount/go-build/util/makefile"

type Builder interface {
	New() Builder
	Link(t *Target) Builder

	// Build the Target's
	Build(makefile.Builder) makefile.Builder

	// End ends building dependencies for a Target
	End() Builder

	// Target adds a new target that will invoke the builder
	Target(target string, dependencies ...string) Builder
	PhonyTarget(target string, dependencies ...string) Builder

	// GetTarget returns the current Target
	GetTarget() *Target

	// GetNamedTarget returns the specified target or nil if not found
	GetNamedTarget(target string) *Target

	BuildTool(flag string, args ...string) Builder
	Echo(label, format string, args ...any) Builder
	Line(format string, args ...any) Builder
	MkDir(path string) Builder

	RootTarget() Builder
}

type builder struct {
	parent   *builder   // Parent
	target   *Target    // Target built
	children []*builder // Child builders
	linked   []*Target  // Linked targets
}

func New() Builder {
	return &builder{}
}

func (b *builder) New() Builder {
	return b.add(nil)
}

func (b *builder) Build(builder makefile.Builder) makefile.Builder {
	// Find the root
	root := b
	//for root.parent != nil {
	//	root = root.parent
	//}

	return root.build(builder)
}

func (b *builder) build(builder makefile.Builder) makefile.Builder {
	if b.target != nil {
		builder = makefile.Of(b.target.Build).Do(builder)
	}

	for _, c := range b.children {
		_ = c.build(builder)
	}

	for _, t := range b.linked {
		builder.AddDependency(t.target)
	}

	return builder
}

func (b *builder) Link(t *Target) Builder {
	b.linked = append(b.linked, t)
	return b
}

func (b *builder) add(target *Target) Builder {
	c := &builder{parent: b, target: target}
	b.children = append(b.children, c)
	return c
}

func (b *builder) End() Builder {
	if b.parent == nil {
		return b
	}
	return b.parent
}

func (b *builder) GetTarget() *Target {
	return b.target
}

func (b *builder) RootTarget() Builder {
	root := b
	for root.parent != nil {
		root = root.parent
	}
	return root
}
