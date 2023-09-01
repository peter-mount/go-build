package target

import (
	"github.com/peter-mount/go-build/util/makefile"
)

// Target defines a custom target in Makefile used by extensions to define rules to make auxiliary files
type Target struct {
	source       string   // Source file/directory
	target       string   // target of the rule, defaults to directory
	dependencies []string // Dependencies specific to this target
	phony        bool     // true to mark target with .PHONY

	children  []*Target // Dependencies in addition to the implicit ones
	TargetDir string    // Target directory, used for Target if it's not defined

	commands []makefile.Handler // List of handlers to generate the lines
}

func (t *Target) Target() string {
	return t.target
}

func (t *Target) Source() string {
	return t.source
}

func (t *Target) Add(cmd makefile.Handler) {
	t.commands = append(t.commands, cmd)
}

func (t *Target) Phony() *Target {
	t.phony = true
	return t
}

func (t *Target) Build(b makefile.Builder) makefile.Builder {

	// Mark the target as phony
	if t.phony {
		b.Phony(t.target)
	}

	// Create the Rule and add its dependencies
	r := b.Rule(t.target, t.dependencies...)

	for _, c := range t.commands {
		_ = c.Do(r)
	}

	// Now it's dependents
	for _, c := range t.children {
		_ = c.Build(r)
		r.AddDependency(c.Target())
	}

	return r
}

func (b *builder) newTarget(target string, dependencies []string) *Target {
	return &Target{
		target:       target,
		dependencies: dependencies,
	}
}

func (b *builder) Target(target string, dependencies ...string) Builder {
	return b.add(b.newTarget(target, dependencies))
}

func (b *builder) PhonyTarget(target string, dependencies ...string) Builder {
	return b.add(b.newTarget(target, dependencies).Phony())
}

func (b *builder) GetNamedTarget(target string) *Target {
	r := b
	for r.parent != nil {
		r = r.parent
	}
	return r.getNamedTarget(target)
}

func (b *builder) getNamedTarget(target string) *Target {
	if b.target != nil && b.target.target == target {
		return b.target
	}
	for _, c := range b.children {
		if t := c.getNamedTarget(target); t != nil {
			return t
		}
	}
	return nil
}
