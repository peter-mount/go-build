package core

import (
	"github.com/peter-mount/go-build/util/jenkinsfile"
	"github.com/peter-mount/go-build/util/meta"
)

type Jenkins func(builder jenkinsfile.Builder, meta *meta.Meta)

func (a Jenkins) Do(builder jenkinsfile.Builder, meta *meta.Meta) {
	if a != nil {
		a(builder, meta)
	}
}

func (a Jenkins) Then(b Jenkins) Jenkins {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return func(builder jenkinsfile.Builder, meta *meta.Meta) {
		a(builder, meta)
		b(builder, meta)
	}
}

func (s *Build) Jenkins(ext Jenkins) {
	s.jenkins = s.jenkins.Then(ext)
}
