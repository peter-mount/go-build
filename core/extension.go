package core

import (
	"github.com/peter-mount/go-build/util/arch"
	"github.com/peter-mount/go-build/util/makefile/target"
	"github.com/peter-mount/go-build/util/meta"
)

// Extension is a hook which is invoked against a target
type Extension func(arch arch.Arch, target target.Builder, meta *meta.Meta)

func (a Extension) Do(arch arch.Arch, target target.Builder, meta *meta.Meta) {
	if a != nil {
		a(arch, target, meta)
	}
}

func (a Extension) Then(b Extension) Extension {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return func(arch arch.Arch, target target.Builder, meta *meta.Meta) {
		a(arch, target, meta)
		b(arch, target, meta)
	}
}

func (s *Build) AddExtension(ext Extension) {
	s.extensions = s.extensions.Then(ext)
}
