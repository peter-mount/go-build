package core

import (
	"github.com/peter-mount/go-build/util/jenkinsfile"
	"sort"
)

type Jenkins func(builder, node jenkinsfile.Builder)

func (a Jenkins) Do(builder, node jenkinsfile.Builder) {
	if a != nil {
		a(builder, node)
	}
}

type JenkinsList []jenkinsEntry

type jenkinsEntry struct {
	seq   int
	entry Jenkins
}

func (l *JenkinsList) Add(seq int, entry Jenkins) {
	*l = append(*l, jenkinsEntry{
		seq:   seq,
		entry: entry,
	})
}

func (l *JenkinsList) IsEmpty() bool {
	return l == nil || len(*l) == 0
}

func (l *JenkinsList) ForEach(f func(Jenkins)) {
	sort.SliceStable(*l, func(i, j int) bool {
		return (*l)[i].seq < (*l)[j].seq
	})

	for _, e := range *l {
		f(e.entry)
	}
}

func (s *Build) Jenkins(seq int, ext Jenkins) {
	s.jenkins.Add(seq, ext)
}
