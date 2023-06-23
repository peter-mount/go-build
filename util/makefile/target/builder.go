package target

import "github.com/peter-mount/go-build/util/makefile"

type Builder interface {
	// Build the Target's
	Build(makefile.Builder) makefile.Builder

	// End ends building dependencies for a Target
	End() Builder

	// Target adds a new target that will invoke the builder
	Target(target string, dependencies ...string) Builder

	// GetTarget returns the current Target
	GetTarget() *Target

	// BuildTool adds a call to the build tool to the Target
	BuildTool(flag string, args ...string) Builder
}

type builder struct {
	parent   *builder   // Parent
	target   *Target    // Target built
	children []*builder // Child builders
}

func New() Builder {
	return &builder{}
}

func (b *builder) Build(builder makefile.Builder) makefile.Builder {
	// Find the root
	root := b
	for root.parent != nil {
		root = root.parent
	}

	// Process the child builders. This is only valid for the root
	// as it's the only one with no target
	for _, c := range root.children {
		_ = makefile.Of(c.target.Build).Do(builder)
	}
	return builder
}

func (b *builder) add(target *Target) Builder {
	c := &builder{parent: b, target: target}

	// Only the root builder has children
	if b.parent == nil {
		b.children = append(b.children, c)
	}

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
