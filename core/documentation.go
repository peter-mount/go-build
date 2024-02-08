package core

import (
	"github.com/peter-mount/go-build/util/makefile"
	"github.com/peter-mount/go-build/util/makefile/target"
	"github.com/peter-mount/go-build/util/meta"
)

// Documentation is a hook which is invoked against a target
type Documentation func(root makefile.Builder, target target.Builder, meta *meta.Meta)

func (a Documentation) Do(root makefile.Builder, target target.Builder, meta *meta.Meta) {
	if a != nil {
		a(root, target, meta)
	}
}

func (a Documentation) Then(b Documentation) Documentation {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return func(root makefile.Builder, target target.Builder, meta *meta.Meta) {
		a(root, target, meta)
		b(root, target, meta)
	}
}

func (s *Build) Documentation(ext Documentation) {
	s.documentation = s.documentation.Then(ext)
}

func (s *Build) Makefile(ext Documentation) {
	s.makefile = s.makefile.Then(ext)
}
