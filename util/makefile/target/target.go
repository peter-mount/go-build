package target

import (
	"github.com/peter-mount/go-build/util/makefile"
)

// Target defines a custom target in Makefile used by extensions to define rules to make auxiliary files
type Target struct {
	source       string   // Source file/directory
	target       string   // target of the rule, defaults to directory
	dependencies []string // Dependencies specific to this target

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

func (t *Target) Build(b makefile.Builder) makefile.Builder {
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

	return b
}

func (b *builder) Target(target string, dependencies ...string) Builder {
	return b.add(&Target{
		target:       target,
		dependencies: dependencies,
	})
}
