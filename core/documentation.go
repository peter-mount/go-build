package core

import (
	"github.com/peter-mount/go-build/util/makefile/target"
	"github.com/peter-mount/go-build/util/meta"
)

// Documentation is a hook which is invoked against a target
type Documentation func(target target.Builder, meta *meta.Meta)

func (a Documentation) Do(target target.Builder, meta *meta.Meta) {
	if a != nil {
		a(target, meta)
	}
}

func (a Documentation) Then(b Documentation) Documentation {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return func(target target.Builder, meta *meta.Meta) {
		a(target, meta)
		b(target, meta)
	}
}

func (s *Build) Documentation(ext Documentation) {
	s.documentation = s.documentation.Then(ext)
}
