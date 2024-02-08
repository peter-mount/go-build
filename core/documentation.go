package core

import (
	"github.com/peter-mount/go-build/util/makefile"
	"github.com/peter-mount/go-build/util/makefile/target"
	"github.com/peter-mount/go-build/util/meta"
	"sort"
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

type DocumentationList []documentationEntry

type documentationEntry struct {
	seq   int
	entry Documentation
}

func (l *DocumentationList) Add(seq int, entry Documentation) {
	*l = append(*l, documentationEntry{
		seq:   seq,
		entry: entry,
	})
}

func (l *DocumentationList) IsEmpty() bool {
	return l == nil || len(*l) == 0
}

func (l *DocumentationList) ForEach(f func(Documentation)) {
	sort.SliceStable(l, func(i, j int) bool {
		return (*l)[i].seq < (*l)[j].seq
	})

	for _, e := range *l {
		f(e.entry)
	}
}

func (s *Build) Documentation(seq int, ext Documentation) {
	s.documentation.Add(seq, ext)
}

func (s *Build) Makefile(seq int, ext Documentation) {
	s.makefile.Add(seq, ext)
}
