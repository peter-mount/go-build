package core

import (
	"github.com/peter-mount/go-build/util/jenkinsfile"
)

type Jenkins func(builder, node jenkinsfile.Builder)

func (a Jenkins) Do(builder, node jenkinsfile.Builder) {
	if a != nil {
		a(builder, node)
	}
}

func (a Jenkins) Then(b Jenkins) Jenkins {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return func(builder, node jenkinsfile.Builder) {
		a(builder, node)
		b(builder, node)
	}
}

func (s *Build) Jenkins(ext Jenkins) {
	s.jenkins = s.jenkins.Then(ext)
}
